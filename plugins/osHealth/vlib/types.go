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
