package stdatem

import (
	"encoding/json"
	"testing"

	"github.com/FlowingSPDG/go-atem"
	"github.com/stretchr/testify/assert"
)

func TestPreviewPropertyInspector_Parse(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		input       string
		expected    *previewPropertyInspector
		expectError bool
	}{
		{
			name:  "正常な設定 - Video Input 1",
			input: `{"ip":"192.168.1.100","input":"1","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoInput1,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Video Input 16",
			input: `{"ip":"192.168.1.100","input":"16","meIndex":"1"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoInput16,
				MeIndex: 1,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Color Bars",
			input: `{"ip":"192.168.1.100","input":"1000","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.ColorBars,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Color 1",
			input: `{"ip":"192.168.1.100","input":"2001","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.Color1,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Media Player 1",
			input: `{"ip":"192.168.1.100","input":"3010","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.MediaPlayer1,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - ME1 Program",
			input: `{"ip":"192.168.1.100","input":"10010","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.ME1Prog,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - ME1 Preview",
			input: `{"ip":"192.168.1.100","input":"10011","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.ME1Prev,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - 最大MeIndex",
			input: `{"ip":"192.168.1.100","input":"1","meIndex":"255"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoInput1,
				MeIndex: 255,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - 未知のInput（VideoBlackに変換）",
			input: `{"ip":"192.168.1.100","input":"99999","meIndex":"0"}`,
			expected: &previewPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoBlack,
				MeIndex: 0,
			},
			expectError: false,
		},

		{
			name:        "失敗 - 無効なInput（小数）",
			input:       `{"ip":"192.168.1.100","input":"1.5","meIndex":"0"}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "失敗 - 無効なMeIndex（小数）",
			input:       `{"ip":"192.168.1.100","input":"1","meIndex":"0.5"}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inspector PreviewPropertyInspector
			err := json.Unmarshal([]byte(tt.input), &inspector)
			if err != nil {
				t.Fatalf("JSONのアンマーシャルに失敗: %v", err)
			}

			result, err := inspector.Parse()

			if tt.expectError {
				asserts.Error(err)
				asserts.Nil(result)
			} else {
				asserts.NoError(err)
				asserts.NotNil(result)
				asserts.Equal(tt.expected.IP, result.IP)
				asserts.Equal(tt.expected.Input, result.Input)
				asserts.Equal(tt.expected.MeIndex, result.MeIndex)
			}
		})
	}
}

func TestProgramPropertyInspector_Parse(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name        string
		input       string
		expected    *programPropertyInspector
		expectError bool
	}{
		{
			name:  "正常な設定 - Video Input 1",
			input: `{"ip":"192.168.1.100","input":"1","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoInput1,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Video Input 16",
			input: `{"ip":"192.168.1.100","input":"16","meIndex":"1"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoInput16,
				MeIndex: 1,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Color Bars",
			input: `{"ip":"192.168.1.100","input":"1000","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.ColorBars,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Color 1",
			input: `{"ip":"192.168.1.100","input":"2001","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.Color1,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - Media Player 1",
			input: `{"ip":"192.168.1.100","input":"3010","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.MediaPlayer1,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - ME1 Program",
			input: `{"ip":"192.168.1.100","input":"10010","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.ME1Prog,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - ME1 Preview",
			input: `{"ip":"192.168.1.100","input":"10011","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.ME1Prev,
				MeIndex: 0,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - 最大MeIndex",
			input: `{"ip":"192.168.1.100","input":"1","meIndex":"255"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoInput1,
				MeIndex: 255,
			},
			expectError: false,
		},
		{
			name:  "正常な設定 - 未知のInput（VideoBlackに変換）",
			input: `{"ip":"192.168.1.100","input":"99999","meIndex":"0"}`,
			expected: &programPropertyInspector{
				IP:      "192.168.1.100",
				Input:   atem.VideoBlack,
				MeIndex: 0,
			},
			expectError: false,
		},

		{
			name:        "失敗 - 無効なInput（小数）",
			input:       `{"ip":"192.168.1.100","input":"1.5","meIndex":"0"}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "失敗 - 無効なMeIndex（小数）",
			input:       `{"ip":"192.168.1.100","input":"1","meIndex":"0.5"}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inspector ProgramPropertyInspector
			err := json.Unmarshal([]byte(tt.input), &inspector)
			if err != nil {
				t.Fatalf("JSONのアンマーシャルに失敗: %v", err)
			}

			result, err := inspector.Parse()

			if tt.expectError {
				asserts.Error(err)
				asserts.Nil(result)
			} else {
				asserts.NoError(err)
				asserts.NotNil(result)
				asserts.Equal(tt.expected.IP, result.IP)
				asserts.Equal(tt.expected.Input, result.Input)
				asserts.Equal(tt.expected.MeIndex, result.MeIndex)
			}
		})
	}
}

func TestSolveATEMVideoInput_Comprehensive(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name     string
		input    int64
		expected atem.VideoInputType
	}{
		// Video Inputs
		{"Video Input 1", 1, atem.VideoInput1},
		{"Video Input 2", 2, atem.VideoInput2},
		{"Video Input 3", 3, atem.VideoInput3},
		{"Video Input 4", 4, atem.VideoInput4},
		{"Video Input 5", 5, atem.VideoInput5},
		{"Video Input 6", 6, atem.VideoInput6},
		{"Video Input 7", 7, atem.VideoInput7},
		{"Video Input 8", 8, atem.VideoInput8},
		{"Video Input 9", 9, atem.VideoInput9},
		{"Video Input 10", 10, atem.VideoInput10},
		{"Video Input 11", 11, atem.VideoInput11},
		{"Video Input 12", 12, atem.VideoInput12},
		{"Video Input 13", 13, atem.VideoInput13},
		{"Video Input 14", 14, atem.VideoInput14},
		{"Video Input 15", 15, atem.VideoInput15},
		{"Video Input 16", 16, atem.VideoInput16},
		{"Video Input 17", 17, atem.VideoInput17},
		{"Video Input 18", 18, atem.VideoInput18},
		{"Video Input 19", 19, atem.VideoInput19},
		{"Video Input 20", 20, atem.VideoInput20},
		
		// Special Sources
		{"Color Bars", 1000, atem.ColorBars},
		{"Color 1", 2001, atem.Color1},
		{"Color 2", 2002, atem.Color2},
		
		// Media Players
		{"Media Player 1", 3010, atem.MediaPlayer1},
		{"Media Player 1 Key", 3011, atem.MediaPlayer1Key},
		{"Media Player 2", 3020, atem.MediaPlayer2},
		{"Media Player 2 Key", 3021, atem.MediaPlayer2Key},
		
		// Key Masks
		{"Key 1 Mask", 4010, atem.Key1Mask},
		{"Key 2 Mask", 4020, atem.Key2Mask},
		{"Key 3 Mask", 4030, atem.Key3Mask},
		{"Key 4 Mask", 4040, atem.Key4Mask},
		
		// DSK Masks
		{"DSK 1 Mask", 5010, atem.DSK1Mask},
		{"DSK 2 Mask", 5020, atem.DSK2Mask},
		
		// Other Sources
		{"Super Source", 6000, atem.SuperSource},
		{"Clean Feed 1", 7001, atem.CleanFeed1},
		{"Clean Feed 2", 7002, atem.CleanFeed2},
		{"Auxiliary 1", 8001, atem.Auxilary1},
		{"Auxiliary 2", 8002, atem.Auxilary2},
		{"Auxiliary 3", 8003, atem.Auxilary3},
		{"Auxiliary 4", 8004, atem.Auxilary4},
		{"Auxiliary 5", 8005, atem.Auxilary5},
		{"Auxiliary 6", 8006, atem.Auxilary6},
		
		// ME Sources
		{"ME1 Program", 10010, atem.ME1Prog},
		{"ME1 Preview", 10011, atem.ME1Prev},
		{"ME2 Program", 10020, atem.ME2Prog},
		{"ME2 Preview", 10021, atem.ME2Prev},
		
		// Edge Cases
		{"Unknown Input", 99999, atem.VideoBlack},
		{"Negative Input", -1, atem.VideoBlack},
		{"Zero Input", 0, atem.VideoBlack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := solveATEMVideoInput(tt.input)
			asserts.Equal(tt.expected, result, "Input: %d", tt.input)
		})
	}
}
