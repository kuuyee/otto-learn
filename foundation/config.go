package foundation

// Config 用来配置各种foundation
type Config struct {
	// Service设置用来配置服务发现
	// (如果有的话)
	//
	// ServiceName 是一个服务的命名。如果是空白
	// 服务发现将不会为应用配置这个服务
	//
	// ServicePort 是服务运行的监听端口
	//
	// ServiceTags 服务打上的一组的Tag
	ServiceName string
	ServicePort int
	ServiceTags []string
}
