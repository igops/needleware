package dynamic

// +k8s:deepcopy-gen=true

type Needleware struct {
	Needles map[string]*Needle `json:"needles,omitempty" toml:"needles,omitempty" yaml:"needles,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

type Needle struct {
	Endpoint        string          `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" export:"true"`
	Client          *NeedleClient   `json:"client,omitempty" toml:"client,omitempty" yaml:"client,omitempty" export:"true"`
	Decision        *NeedleDecision `json:"decision,omitempty" toml:"decision,omitempty" yaml:"decision,omitempty" export:"true"`
	NotifyConnClose []string        `json:"notifyConnClose,omitempty" toml:"notifyConnClose,omitempty" yaml:"notifyConnClose,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

type NeedleDecision struct {
	OnTimeout string `json:"onTimeout,omitempty" toml:"onTimeout,omitempty" yaml:"onTimeout,omitempty" export:"true"`
	OnError   string `json:"onError,omitempty" toml:"onError,omitempty" yaml:"onError,omitempty" export:"true"`
	OnReject  string `json:"onReject,omitempty" toml:"onReject,omitempty" yaml:"onReject,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

type NeedleClient struct {
	Type    string      `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty" export:"true"`
	Timeout string      `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	Auth    *NeedleAuth `json:"auth,omitempty" toml:"auth,omitempty" yaml:"auth,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

type NeedleAuth struct {
	Method          string `json:"method,omitempty" toml:"method,omitempty" yaml:"method,omitempty" export:"true"`
	TlsCertFilePath string `json:"tlsCertFilePath,omitempty" toml:"tlsCertFilePath,omitempty" yaml:"tlsCertFilePath,omitempty" export:"true"`
}

/**
needleware:
	needles:
		some-needle:
			endpoint: "localhost:50051"
			client:
				type: grpc
				timeout: 10
				auth:
					method: tls
					tlsCertFilePath: "/path/to/cert"
			decision:
				onTimeout: reject | accept
				onError: reject | accept
			notifyConnClose:
				- accept
				- reject
*/
