package models

// CacheSubData cache 보조 데이터
type CacheSubData struct {
	DSL         string            `json:"-"`
	IsDateRange bool              `json:"isDateRange"`
	Value       CacheSubDataValue `json:"value"`
}

// CacheSubDataValue cache 보조 데이터 바디(value)
type CacheSubDataValue struct {
	DataModels     []*DataModel      `json:"dataModels"`
	DataSourceInfo []*DataSourceInfo `json:"dataSourceInfo"`
}

// DataSet IRIS 데이터 모델 feild중 dataset의 JSON 포맷
type DataSet struct {
	Format string `json:"format"`
	Table  string `json:"table"`
	Path   string `json:"path"`
	Query  string `json:"query"`
}

// DataModel IRIS 데이터 모델
type DataModel struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Dataset        DataSet `json:"dataset"`
	ConnectorID    string  `json:"connectorID"`
	Description    string  `json:"description"`
	Fields         string  `json:"fields"`
	Share          string  `json:"share"`
	UserID         string  `json:"userID"`
	GroupID        string  `json:"groupID"`
	CDate          string  `json:"CDate"`
	MDate          string  `json:"MDate"`
	Scope          string  `json:"scope"`
	PartitionRange string  `json:"partitionRange"`
	ReferencedID   string  `json:"referencedID"`
}

// DataSourceInfo 데이터 연결 정보, 테이블(파일) / 데이터의 변경 정보
type DataSourceInfo struct {
	DataSourceConnInfo *DataSourceConnInfo       `json:"dataSourceConnInfo"`
	LastModifyInfo     *DataSourceLastModifyInfo `json:"lastModifyInfo"`
}

// DataSourceConnInfo 원본 데이터 저장소의 연결 정보와 데이터가 저장된 테이블(파일) 정보
type DataSourceConnInfo struct {
	DataStoreType string `json:"dataStoreType"` // 데이터 저장소 타입
	IP            string `json:"ip"`            // 연결 정보
	Port          int    `json:"port"`
	User          string `json:"user"` // 인증 정보
	Password      string `json:"password"`
	AccessKey     string `json:"accessKey"` // for minio
	SecretKey     string `json:"secretKey"`
	SSL           bool   `json:"ssl"`
	DatabaseName  string `json:"databaseName"` // mariadb, postgresql
	SchemaName    string `json:"schemaName"`
	TableName     string `json:"tableName"`
	BucketName    string `json:"bucketName"`   // Minio
	AbsolutePath  string `json:"absolutePath"` // Minio, HDFS
}

// DataSourceLastModifyInfo for orignal data source modify info
type DataSourceLastModifyInfo struct {
	DataSourceType string      `json:"dataSourceType"`
	Info           interface{} `json:"info"`
}
