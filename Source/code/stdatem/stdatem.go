package stdatem

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/FlowingSPDG/go-atem"
	"github.com/FlowingSPDG/std-atem/Source/code/connectionmanager"
	"github.com/FlowingSPDG/std-atem/Source/code/logger"
	"github.com/FlowingSPDG/std-atem/Source/code/setting"
	"github.com/FlowingSPDG/streamdeck"
	sdcontext "github.com/FlowingSPDG/streamdeck/context"
)

// App メインエンジン
type App struct {
	connectionManager   *connectionmanager.ConnectionManager // コンテキスト（ボタン）ごとの設定
	logger              logger.Logger                        // ログ
	sd                  *streamdeck.Client                   // StreamDeckクライアント
	previewSettingStore setting.SettingStore[*previewPropertyInspector]
	programSettingStore setting.SettingStore[*programPropertyInspector]
	// refCounts and activeClients removed - now managed by ConnectionManager
}

// NewApp Appメインエンジンを初期化する
func NewApp(ctx context.Context, logger logger.Logger, sd *streamdeck.Client) (*App, error) {
	app := &App{
		connectionManager:   connectionmanager.NewConnectionManager(logger),
		logger:              logger,
		sd:                  sd,
		previewSettingStore: setting.NewSettingStore[*previewPropertyInspector](),
		programSettingStore: setting.NewSettingStore[*programPropertyInspector](),
	}

	app.setupSD()

	return app, nil
}

// recomputeTallies 指定IPに紐づく全コンテキストのタリーを再計算
func (a *App) recomputeTallies(ctx context.Context, ip string, instance *connectionmanager.ATEMInstance) {
	// 紐づいたContextを取得
	actions, ok := a.connectionManager.SolveContextsByIP(ctx, ip)
	if !ok {
		a.logger.Error(ctx, "recomputeTallies ATEMが見つかりません")
		return
	}

	for _, ac := range actions {
		// 設定を解決（preview/program どちらの設定でもInputを取得できればOK）
		var (
			found    bool
			input    atem.VideoInputType
			tallyPRV bool = true
			tallyPGM bool = true
		)
		if s, ok := a.previewSettingStore.Load(ac.Context); ok {
			input = s.Input
			tallyPRV = s.TallyPRV
			tallyPGM = s.TallyPGM
			found = true
		} else if s, ok := a.programSettingStore.Load(ac.Context); ok {
			input = s.Input
			tallyPRV = s.TallyPRV
			tallyPGM = s.TallyPGM
			found = true
		}
		if !found {
			continue
		}

		// 現在のPGM/PRVと比較して画像を設定（フラグに応じて適用）
		sdctx := sdcontext.WithContext(ctx, ac.Context)
		matchPGM := uint8(input) == uint8(instance.Client.ProgramInput.Index) && tallyPGM
		matchPRV := uint8(input) == uint8(instance.Client.PreviewInput.Index) && tallyPRV
		switch {
		case matchPGM:
			a.sd.SetImage(sdctx, tallyProgram, streamdeck.HardwareAndSoftware)
		case matchPRV:
			a.sd.SetImage(sdctx, tallyPreview, streamdeck.HardwareAndSoftware)
		default:
			a.sd.SetImage(sdctx, tallyInactive, streamdeck.HardwareAndSoftware)
		}
	}
}

// addATEMHost 新しいATEMホストを追加し、接続をセットアップする
func (a *App) addATEMHost(ctx context.Context, action string, contextID string, ip string, debug bool) error {
	msg := fmt.Sprintf("ATEMホスト %s を追加中...", ip)
	a.logger.Debug(ctx, msg)

	// 空IPや不正なIPは無視（PI入力中のスパム防止）
	if ip == "" {
		return nil
	}
	if parsed := net.ParseIP(ip); parsed == nil || parsed.To4() == nil {
		return nil
	}

	if instance, ok := a.connectionManager.SolveATEMByIP(ctx, ip); ok {
		a.logger.Debug(ctx, "ATEMホスト %s は既に存在します", ip)
		// 既存接続を再利用し、新しいcontextを追加
		a.connectionManager.Store(ctx, action, ip, contextID, instance)
		return nil
	}

	instance := &connectionmanager.ATEMInstance{
		Client:      atem.Create(ip, debug),
		ReconnectCh: make(chan struct{}, 1),
	}

	a.connectionManager.Store(ctx, action, ip, contextID, instance)

	instance.Client.On("connected", func() {
		a.logger.Debug(ctx, fmt.Sprintf("ATEM %s に接続しました", ip))
	})

	instance.Client.On("PrvI.change", func() {
		a.recomputeTallies(ctx, ip, instance)
	})

	instance.Client.On("PrgI.change", func() {
		a.recomputeTallies(ctx, ip, instance)
	})

	instance.Client.On("closed", func() {
		a.logger.Debug(ctx, fmt.Sprintf("ATEM %s への接続を閉じました", ip))
		if instance, ok := a.connectionManager.SolveATEMByIP(ctx, ip); ok {

			// 再接続をトリガー
			select {
			case instance.ReconnectCh <- struct{}{}:
				a.logger.Debug(ctx, "reconnectionLoop ip:%s 再接続をトリガーしました", ip)
			case <-ctx.Done():
				return
			default:
			}
		}
	})

	// 再接続ゴルーチンを開始（まだ実行されていない場合のみ）
	if ip != "" && !a.connectionManager.IsReconnectGoroutineRunning(ip) {
		a.connectionManager.SetReconnectGoroutineRunning(ip, true)
		go a.reconnectionLoop(ctx, ip)
		a.logger.Debug(ctx, "addATEMHost ip:%s 再接続ゴルーチンを開始", ip)
	}
	if ip != "" {
		instance.ReconnectCh <- struct{}{}
	}

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
	defer func() {
		// ゴルーチン終了時に状態をクリア
		a.connectionManager.SetReconnectGoroutineRunning(ip, false)
		a.logger.Debug(ctx, "reconnectionLoop ip:%s ゴルーチンを終了", ip)
	}()

	instance, ok := a.connectionManager.SolveATEMByIP(ctx, ip)
	if !ok {
		a.logger.Error(ctx, "ATEMが見つかりません")
		return
	}

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			a.logger.Debug(ctx, "reconnectionLoop ip:%s コンテキストが終了したため終了", ip)
			return
		case <-instance.ReconnectCh:
			a.logger.Debug(ctx, "reconnectionLoop ip:%s 再接続をトリガーしました", ip)
			if err := instance.Client.Connect(); err != nil {
				// 指数バックオフ: 1秒, 2秒, 4秒, 8秒, 16秒, 30秒(最大)
				backoff = min(backoff*2, maxBackoff)
				a.logger.Debug(ctx, "reconnectionLoop ip:%s 接続失敗、%v後に再試行", ip, backoff)

				// 再試行前に待機
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}

				// 再試行
				select {
				case instance.ReconnectCh <- struct{}{}:
				default:
				}
			} else {
				// 接続成功時はバックオフをリセット
				backoff = time.Second
				a.logger.Debug(ctx, "reconnectionLoop ip:%s 接続成功、バックオフをリセット", ip)
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

func (a *App) handleDisappear(ctx context.Context, contextID string) {
	a.logger.Debug(ctx, "handleDisappear contextID:%s", contextID)
	a.connectionManager.DeleteATEMByContext(ctx, contextID)
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
