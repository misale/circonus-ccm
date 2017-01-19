package ccm

//circonus tenant info struct
type circonusTenant struct {
	API_URL   string `yaml:"circonus_api_url"`
	API_TOKEN string `yaml:"circonus_api_token"`
	APP_NAME  string `yaml:"circonus_app_name"`
}

//Host group struct
type HostGroup struct {
	GroupName string   `yaml:"group_name"`
	Members   []string `yaml:"members"`
}

// Check Input struct for creating a check bundle
type CheckInput struct {
	Brokers []string `json:"brokers"`
	Config  struct {
		HeaderHost  string `json:"header_host"`
		HTTPVersion string `json:"http_version"`
		Method      string `json:"method"`
		Payload     string `json:"payload"`
		Port        string `json:"port"`
		ReadLimit   string `json:"read_limit"`
		URL         string `json:"url"`
		Query       string `json:"query"`
	} `json:"config"`
	DisplayName string   `json:"display_name"`
	Notes       string   `json:"notes"`
	Period      float64  `json:"period"`
	Tags        []string `json:"tags"`
	Target      string   `json:"target"`
	Timeout     float64  `json:"timeout"`
	Type        string   `json:"type"`
	Metrics     []struct {
		Status string      `json:"status"`
		Name   string      `json:"name"`
		Type   string      `json:"type"`
		Units  interface{} `json:"units"`
		Tags   []string    `json:"tags"`
	} `json:"metrics"`
}

// ccm template file struct
type CcmTemplate struct {
	HostGroup         string `json:"host_group"`
	TemplateFile      string `json:"template_file"`
	Broker            string `json:"broker"`
	DisplayName       string `json:"display_name"`
	Notes             string `json:"notes"`
	Period            int    `json:"period"`
	Target            string `json:"target"`
	Timeout           int    `json:"timeout"`
	Type              string `json:"type"`
	ConfigHeaderHost  string `json:"config_header_host"`
	ConfigHTTPVersion string `json:"config_http_version"`
	ConfigMethod      string `json:"config_method"`
	ConfigPayload     string `json:"config_payload"`
	ConfigPort        string `json:"config_port"`
	ConfigReadLimit   string `json:"config_read_limit"`
	ConfigURL         string `json:"config_url"`
	ConfigQuery       string `json:"config_query"`
}

//ccm conf file struct (exposing only hostgroup and template file name fields)
type CCMConf struct {
	HostGroup    string `json:"host_group"`
	TemplateFile string `json:"template_file"`
}
