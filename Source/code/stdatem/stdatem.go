package stdatem

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/FlowingSPDG/std-atem/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
	"github.com/puzpuzpuz/xsync"
	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

// App メインエンジン
type App struct {
	atems               *atems             // コンテキスト（ボタン）ごとの設定
	logger              logger.Logger      // ログ
	sd                  *streamdeck.Client // StreamDeckクライアント
	previewSettingStore setting.SettingStore[*previewPropertyInspector]
	programSettingStore setting.SettingStore[*programPropertyInspector]
	refCounts           *xsync.MapOf[string, int]
	activeClients       *xsync.MapOf[string, *ATEMInstance]
}

// NewApp Appメインエンジンを初期化する
func NewApp(ctx context.Context, logger logger.Logger, sd *streamdeck.Client) (*App, error) {
	app := &App{
		atems:               newAtems(logger),
		logger:              logger,
		sd:                  sd,
		previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
		programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
		refCounts:           xsync.NewMapOf[int](),
		activeClients:       xsync.NewMapOf[*ATEMInstance](),
	}

	// SDのセットアップ
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return nil, xerrors.Errorf("registration paramsの解析に失敗: %w", err)
	}
	app.sd = streamdeck.NewClient(ctx, params)
	app.setupSD()

	return app, nil
}

// addATEMHost 新しいATEMホストを追加し、接続をセットアップする
func (a *App) addATEMHost(ctx context.Context, action string, contextID string, ip string, debug bool) error {
	msg := fmt.Sprintf("ATEMホスト %s を追加中...", ip)
	a.logger.Debug(ctx, msg)

	if instance, ok := a.atems.atemByIP.Load(ip); ok {
		a.logger.Debug(ctx, "ATEMホスト %s は既に存在します", ip)
		contexts, ok := a.atems.SolveContextsByIP(ctx, ip)
		if !ok {
			a.logger.Error(ctx, "ATEMが見つかりません")
			return xerrors.New("ATEMが見つかりません")
		} else {
			contexts = append(contexts, actionAndContext{action: action, context: contextID})
			a.atems.contextsByIP.Store(ip, contexts)
			a.atems.Store(ctx, action, ip, contextID, instance)
			return nil
		}
	}

	instance := &ATEMInstance{
		client:      atem.Create(ip, debug),
		reconnectCh: make(chan struct{}, 1),
	}

	a.atems.Store(ctx, action, ip, contextID, instance)

	instance.client.On("connected", func() {
		a.logger.Debug(ctx, fmt.Sprintf("ATEM %s に接続しました", ip))
	})

	instance.client.On("PrvI.change", func() {
		a.logger.Debug(ctx, "PrvI.change")

		// 紐づいたContextを取得
		actions, ok := a.atems.SolveContextsByIP(ctx, ip)
		if !ok {
			a.logger.Error(ctx, "PrvI.change ATEMが見つかりません")
			return
		}
		a.logger.Debug(ctx, "PrvI.change actions:%v", actions)
		actions = lo.Filter(actions, func(action actionAndContext, _ int) bool {
			return action.action == setPreviewAction
		})
		a.logger.Debug(ctx, "PrvI.change contexts:%v", actions)

		for _, action := range actions {
			previewSetting, ok := a.previewSettingStore.Load(action.context)
			if !ok {
				a.logger.Error(ctx, "previewSettingが見つかりません")
				return
			}

			// TODO: M/Eが違う場合は無視する
			a.logger.Debug(ctx, "PrvI.change input:%d meIndex:%d PreviewInput:%v", previewSetting.Input, previewSetting.MeIndex, instance.client.PreviewInput)
			isActive := uint8(previewSetting.Input) == uint8(instance.client.PreviewInput.Index)
			a.logger.Debug(ctx, "PrvI.change setting:%v actual:%d isActive:%t", previewSetting, instance.client.PreviewInput.Index, isActive)

			// タリーを反映
			sdctx := sdcontext.WithContext(ctx, action.context)
			if isActive {
				a.sd.SetImage(sdctx, tallyPreview, streamdeck.HardwareAndSoftware)
			} else {
				a.sd.SetImage(sdctx, tallyInactive, streamdeck.HardwareAndSoftware)
			}
		}
	})

	instance.client.On("PrgI.change", func() {
		a.logger.Debug(ctx, "PrgI.change")

		// 紐づいたContextを取得
		actions, ok := a.atems.SolveContextsByIP(ctx, ip)
		if !ok {
			a.logger.Error(ctx, "PrgI.change ATEMが見つかりません")
			return
		}
		a.logger.Debug(ctx, "PrgI.change actions:%v", actions)
		actions = lo.Filter(actions, func(action actionAndContext, _ int) bool {
			return action.action == setProgramAction
		})
		a.logger.Debug(ctx, "PrgI.change contexts:%v", actions)

		for _, action := range actions {
			programSetting, ok := a.programSettingStore.Load(action.context)
			if !ok {
				a.logger.Error(ctx, "PrgI.change programSettingが見つかりません")
				return
			}

			// TODO: M/Eが違う場合は無視する
			a.logger.Debug(ctx, "PrgI.change input:%d meIndex:%d PreviewInput:%v", programSetting.Input, programSetting.MeIndex, instance.client.ProgramInput.Index)
			isActive := uint8(programSetting.Input) == uint8(instance.client.ProgramInput.Index)
			a.logger.Debug(ctx, "PrgI.change setting:%v actual:%d isActive:%t", programSetting, instance.client.ProgramInput.Index, isActive)

			// タリーを反映
			sdctx := sdcontext.WithContext(ctx, action.context)
			if isActive {
				a.sd.SetImage(sdctx, tallyProgram, streamdeck.HardwareAndSoftware)
			} else {
				a.sd.SetImage(sdctx, tallyInactive, streamdeck.HardwareAndSoftware)
			}

		}
	})

	instance.client.On("closed", func() {
		a.logger.Debug(ctx, fmt.Sprintf("ATEM %s への接続を閉じました", ip))
		if instance, ok := a.atems.SolveATEMByIP(ctx, ip); ok {

			// 再接続をトリガー
			select {
			case instance.reconnectCh <- struct{}{}:
				a.logger.Debug(ctx, "reconnectionLoop ip:%s 再接続をトリガーしました", ip)
			case <-ctx.Done():
				return
			default:
			}
		}
	})

	// 再接続ゴルーチンを開始
	go a.reconnectionLoop(ctx, ip)
	a.logger.Debug(ctx, "addATEMHost ip:%s 再接続ゴルーチンを開始", ip)
	instance.reconnectCh <- struct{}{}

	return nil
}

// Run バックグラウンドプロセスを実行
func (a *App) Run(ctx context.Context) error {
	return a.sd.Run(ctx)
}

// setupSD StreamDeckクライアントをセットアップ
func (a *App) setupSD() {
	setPreviewAction := a.sd.Action(setPreviewAction)
	setPreviewAction.RegisterHandler(streamdeck.KeyDown, a.PRVKeyDownHandler)
	setPreviewAction.RegisterHandler(streamdeck.WillAppear, a.PRVWillAppearHandler)
	setPreviewAction.RegisterHandler(streamdeck.WillDisappear, a.PRVWillDisappearHandler)
	setPreviewAction.RegisterHandler(streamdeck.DidReceiveSettings, a.PRVDidReceiveSettingsHandler)

	setProgramAction := a.sd.Action(setProgramAction)
	setProgramAction.RegisterHandler(streamdeck.KeyDown, a.PGMKeyDownHandler)
	setProgramAction.RegisterHandler(streamdeck.WillAppear, a.PGMWillAppearHandler)
	setProgramAction.RegisterHandler(streamdeck.WillDisappear, a.PGMWillDisappearHandler)
	setProgramAction.RegisterHandler(streamdeck.DidReceiveSettings, a.PGMDidReceiveSettingsHandler)

	cutAction := a.sd.Action(cutAction)
	cutAction.RegisterHandler(streamdeck.KeyDown, a.CutKeyDownHandler)
	cutAction.RegisterHandler(streamdeck.WillAppear, a.CutWillAppearHandler)
	cutAction.RegisterHandler(streamdeck.WillDisappear, a.CutWillDisappearHandler)
	cutAction.RegisterHandler(streamdeck.DidReceiveSettings, a.CutDidReceiveSettingsHandler)

	autoAction := a.sd.Action(autoAction)
	autoAction.RegisterHandler(streamdeck.KeyDown, a.AutoKeyDownHandler)
	autoAction.RegisterHandler(streamdeck.WillAppear, a.AutoWillAppearHandler)
	autoAction.RegisterHandler(streamdeck.WillDisappear, a.AutoWillDisappearHandler)
	autoAction.RegisterHandler(streamdeck.DidReceiveSettings, a.AutoDidReceiveSettingsHandler)

}

// reconnectionLoop 特定のATEMホストの自動再接続を処理
func (a *App) reconnectionLoop(ctx context.Context, ip string) {
	a.logger.Debug(ctx, "reconnectionLoop ip:%s", ip)
	instance, ok := a.atems.SolveATEMByIP(ctx, ip)
	if !ok {
		a.logger.Error(ctx, "ATEMが見つかりません")
		return
	}

	for {
		select {
		case <-ctx.Done():
			a.logger.Debug(ctx, "reconnectionLoop ip:%s コンテキストが終了したため終了", ip)
			return
		case <-instance.reconnectCh:
			a.logger.Debug(ctx, "reconnectionLoop ip:%s 再接続をトリガーしました", ip)
			if err := instance.client.Connect(); err != nil {
				// 再試行前に待機
				time.Sleep(5 * time.Second)
				// 再試行
				select {
				case instance.reconnectCh <- struct{}{}:
				default:
				}
			}
		}
	}
}

func solveATEMVideoInput(input int64) atem.VideoInputType {
	switch input {
	case 1:
		return atem.VideoInput1
	case 2:
		return atem.VideoInput2
	case 3:
		return atem.VideoInput3
	case 4:
		return atem.VideoInput4
	case 5:
		return atem.VideoInput5
	case 6:
		return atem.VideoInput6
	case 7:
		return atem.VideoInput7
	case 8:
		return atem.VideoInput8
	case 9:
		return atem.VideoInput9
	case 10:
		return atem.VideoInput10
	case 11:
		return atem.VideoInput11
	case 12:
		return atem.VideoInput12
	case 13:
		return atem.VideoInput13
	case 14:
		return atem.VideoInput14
	case 15:
		return atem.VideoInput15
	case 16:
		return atem.VideoInput16
	case 17:
		return atem.VideoInput17
	case 18:
		return atem.VideoInput18
	case 19:
		return atem.VideoInput19
	case 20:
		return atem.VideoInput20
	case 1000:
		return atem.ColorBars
	case 2001:
		return atem.Color1
	case 2002:
		return atem.Color2
	case 3010:
		return atem.MediaPlayer1
	case 3011:
		return atem.MediaPlayer1Key
	case 3020:
		return atem.MediaPlayer2
	case 3021:
		return atem.MediaPlayer2Key
	case 4010:
		return atem.Key1Mask
	case 4020:
		return atem.Key2Mask
	case 4030:
		return atem.Key3Mask
	case 4040:
		return atem.Key4Mask
	case 5010:
		return atem.DSK1Mask
	case 5020:
		return atem.DSK2Mask
	case 6000:
		return atem.SuperSource
	case 7001:
		return atem.CleanFeed1
	case 7002:
		return atem.CleanFeed2
	case 8001:
		return atem.Auxilary1
	case 8002:
		return atem.Auxilary2
	case 8003:
		return atem.Auxilary3
	case 8004:
		return atem.Auxilary4
	case 8005:
		return atem.Auxilary5
	case 8006:
		return atem.Auxilary6
	case 10010:
		return atem.ME1Prog
	case 10011:
		return atem.ME1Prev
	case 10020:
		return atem.ME2Prog
	case 10021:
		return atem.ME2Prev
	default:
		return atem.VideoBlack
	}
}

func (a *App) handleDisappear(ctx context.Context, hostname string) {
	a.logger.Debug(ctx, "handleDisappear hostname:%s", hostname)
	a.atems.DeleteATEMByContext(ctx, hostname)
}

// Run アプリケーションを初期化して実行
func Run(ctx context.Context, logger logger.Logger, sd *streamdeck.Client) error {
	// アプリケーションを初期化
	app, err := NewApp(ctx, logger, sd)
	if err != nil {
		return fmt.Errorf("アプリの初期化に失敗: %w", err)
	}

	// アプリケーションを実行
	return app.Run(ctx)
}
