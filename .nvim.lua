vim.lsp.config("gopls", {
	settings = {
		gopls = {
			buildFlags = {
				"-tags=osHealth,ufwApply,linux,windows,darwin,freebsd,mysqlHealth,mariadbHealth,pgsqlHealth",
			},
		},
	},
})
