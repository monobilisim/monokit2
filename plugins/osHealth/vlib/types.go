package vlib

type DockerVersion struct {
	Client struct {
		Version    string
		APIVersion string
		GoVersion  string
		GitCommit  string
		Built      string
		OSArch     string
		Context    string
	}

	Server struct {
		Engine struct {
			Version      string
			APIVersion   string
			GoVersion    string
			GitCommit    string
			Built        string
			OSArch       string
			Experimental bool
		}

		Containerd struct {
			Version   string
			GitCommit string
		}

		Runc struct {
			Version   string
			GitCommit string
		}

		TiniStatic struct {
			Version   string
			GitCommit string
		}
	}
}

type CaddyVersion struct {
	Version     string
	VersionFull string
}

type AsteriskVersion struct {
	Version     string
	VersionFull string
}

type FrankenPHPVersion struct {
	FrankenPHP struct {
		Version     string
		VersionFull string
	}
	PHP struct {
		Version     string
		VersionFull string
	}
	Caddy struct {
		Version     string
		VersionFull string
	}
	VersionFull string
}

type HAProxyVersion struct {
	Version     string
	Status      string
	KnownBugs   string
	RunningOn   string
	VersionFull string
}

type JenkinsVersion struct {
	Version     string
	VersionFull string
}

type MongoDBVersion struct {
	Environment struct {
		Distmod    string `json:"distmod"`
		Distarch   string `json:"distarch"`
		TargetArch string `json:"target_arch"`
	} `json:"environment"`
	Version        string   `json:"version"`
	VersionFull    string   `json:"-"`
	GitVersion     string   `json:"gitVersion"`
	OpenSSLVersion string   `json:"openSSLVersion"`
	Modules        []string `json:"modules"`
	Allocator      string   `json:"allocator"`
}

type MySQLVersion struct {
	Version     string
	VersionFull string
}

type MariaDBVersion struct {
	Version     string
	VersionFull string
}

type NginxVersion struct {
	Version     string
	VersionFull string
}

type OPNsenseVersion struct {
	Version     string
	VersionFull string
}

type PostalVersion struct {
	Version     string
	VersionFull string
}

type PostgreSQLVersion struct {
	Version     string
	VersionFull string
}

type RedisVersion struct {
	Version     string
	VersionFull string
}

type ValkeyVersion struct {
	Version     string
	VersionFull string
}
