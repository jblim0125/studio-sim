package appdata

// 빌드 타임에 설정되는 변수 들
// ldflags -X 옵션으로 설정 됨
var (
	Name      string = "golang-template"
	Version   string = "v1.0.0"
	BuildHash string = "hash1234"
)

// VersionInfo app version info
type VersionInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	BuildHash string `json:"buildHash"`
}
