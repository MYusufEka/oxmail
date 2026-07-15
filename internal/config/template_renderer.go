package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates
var templatesFS embed.FS

// RenderPayload holds data injected into all .tmpl config files.
type RenderPayload struct {
	Hostname         string
	Domain           string
	MessageSizeLimit string
	DevMode          bool
}

// RenderAll renders all .tmpl config files to the given output directory.
//   - postfix/main.cf, postfix/master.cf → outputDir/postfix/
//   - dovecot/dovecot.conf → outputDir/dovecot/
//   - dovecot/10-auth.conf → outputDir/dovecot/conf.d/
//   - dovecot/10-mail.conf → outputDir/dovecot/conf.d/
func RenderAll(outputDir string, data RenderPayload) error {
	type templateFile struct {
		src string // path within templatesFS
		dst string // relative to outputDir
	}

	files := []templateFile{
		{src: "templates/postfix/main.cf.tmpl", dst: "postfix/main.cf"},
		{src: "templates/postfix/master.cf.tmpl", dst: "postfix/master.cf"},
		{src: "templates/dovecot/dovecot.conf.tmpl", dst: "dovecot/dovecot.conf"},
		{src: "templates/dovecot/10-auth.conf.tmpl", dst: "dovecot/conf.d/10-auth.conf"},
		{src: "templates/dovecot/10-mail.conf.tmpl", dst: "dovecot/conf.d/10-mail.conf"},
	}

	for _, f := range files {
		dstPath := filepath.Join(outputDir, f.dst)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", f.dst, err)
		}

		tmpl, err := template.ParseFS(templatesFS, f.src)
		if err != nil {
			return fmt.Errorf("parse %s: %w", f.src, err)
		}

		out, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", f.dst, err)
		}
		defer out.Close()

		if err := tmpl.Execute(out, data); err != nil {
			return fmt.Errorf("render %s: %w", f.dst, err)
		}
	}

	return nil
}
