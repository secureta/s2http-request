package functions

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// DictionaryLoadFunction は外部ファイルから辞書データを読み込む関数
type DictionaryLoadFunction struct{}

func (f *DictionaryLoadFunction) Name() string {
	return "dict_load"
}

func (f *DictionaryLoadFunction) Signature() string {
	return "dict_load(file_path: string) -> []string"
}

func (f *DictionaryLoadFunction) Description() string {
	return "外部ファイルから辞書データを読み込みます。対応形式: JSON, YAML, テキスト"
}

func (f *DictionaryLoadFunction) Execute(ctx context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dict_load requires exactly 1 argument, got %d", len(args))
	}

	filePath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("dict_load: file path must be a string")
	}

	// コンテキストからリクエストファイルのパスを取得
	requestFilePath, ok := ctx.Value("requestFilePath").(string)
	if ok && !filepath.IsAbs(filePath) {
		baseDir := filepath.Dir(requestFilePath)
		filePath = filepath.Join(baseDir, filePath)
	}

	// ファイルの存在確認
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("dict_load: file not found: %s", filePath)
	}

	// ファイル拡張子による処理の分岐
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return f.loadJSONFile(filePath)
	case ".yaml", ".yml":
		return f.loadYAMLFile(filePath)
	case ".txt", "":
		return f.loadTextFile(filePath)
	default:
		return nil, fmt.Errorf("dict_load: unsupported file format: %s", ext)
	}
}

func (f *DictionaryLoadFunction) loadJSONFile(filePath string) (result interface{}, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}()

	var data interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON file: %w", err)
	}

	return f.normalizeToStringArray(data)
}

// loadYAMLFile はYAMLファイルを読み込む
func (f *DictionaryLoadFunction) loadYAMLFile(filePath string) (result interface{}, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}()

	var data interface{}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	return f.normalizeToStringArray(data)
}

// loadTextFile はテキストファイルを読み込む（1行1エントリ）
func (f *DictionaryLoadFunction) loadTextFile(filePath string) (result interface{}, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open text file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" { // 空行のみをスキップ
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	return lines, nil
}

// normalizeToStringArray はデータを文字列配列に正規化
func (f *DictionaryLoadFunction) normalizeToStringArray(data interface{}) ([]string, error) {
	switch v := data.(type) {
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result, nil
	case []string:
		return v, nil
	case map[string]interface{}:
		// マップの場合は値のみを取得
		var result []string
		for _, value := range v {
			// 値が配列の場合は再帰的に処理
			if subArray, err := f.normalizeToStringArray(value); err == nil {
				result = append(result, subArray...)
			} else {
				result = append(result, fmt.Sprintf("%v", value))
			}
		}
		return result, nil
	case string:
		return []string{v}, nil
	default:
		return []string{fmt.Sprintf("%v", v)}, nil
	}
}

// DictionaryRandomFunction は辞書からランダムに値を選択する関数
type DictionaryRandomFunction struct{}

func (f *DictionaryRandomFunction) Name() string {
	return "dict_random"
}

func (f *DictionaryRandomFunction) Signature() string {
	return "dict_random(dictionary: []string) -> string"
}

func (f *DictionaryRandomFunction) Description() string {
	return "辞書配列からランダムに1つの値を選択します"
}

func (f *DictionaryRandomFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dict_random requires exactly 1 argument, got %d", len(args))
	}

	switch dict := args[0].(type) {
	case []interface{}:
		if len(dict) == 0 {
			return "", nil
		}
		// Go 1.20以降の推奨方法
		index := rand.Intn(len(dict))
		return fmt.Sprintf("%v", dict[index]), nil
	case []string:
		if len(dict) == 0 {
			return "", nil
		}
		// Go 1.20以降の推奨方法
		index := rand.Intn(len(dict))
		return dict[index], nil
	default:
		return nil, fmt.Errorf("dict_random: argument must be an array")
	}
}

// DictionaryGetFunction は辞書から指定されたインデックスの値を取得する関数
type DictionaryGetFunction struct{}

func (f *DictionaryGetFunction) Name() string {
	return "dict_get"
}

func (f *DictionaryGetFunction) Signature() string {
	return "dict_get(dictionary: []string, index: number) -> string"
}

func (f *DictionaryGetFunction) Description() string {
	return "辞書配列から指定されたインデックスの値を取得します"
}

func (f *DictionaryGetFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("dict_get requires exactly 2 arguments, got %d", len(args))
	}

	var dict []string
	switch d := args[0].(type) {
	case []interface{}:
		dict = make([]string, len(d))
		for i, item := range d {
			dict[i] = fmt.Sprintf("%v", item)
		}
	case []string:
		dict = d
	default:
		return nil, fmt.Errorf("dict_get: first argument must be an array")
	}

	var index int
	switch i := args[1].(type) {
	case int:
		index = i
	case float64:
		index = int(i)
	default:
		return nil, fmt.Errorf("dict_get: second argument must be a number")
	}

	if index < 0 || index >= len(dict) {
		return "", fmt.Errorf("dict_get: index %d out of range [0, %d)", index, len(dict))
	}

	return dict[index], nil
}
