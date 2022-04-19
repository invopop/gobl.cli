package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl.cli/internal"
	"github.com/invopop/gobl/dsig"
)

type buildOpts struct {
	*rootOpts
	envelop        bool // when true, assumes source is a document
	set            map[string]string
	setFiles       map[string]string
	setStrings     map[string]string
	template       string
	privateKeyFile string
	docType        string

	// Command options
	use   string
	short string
}

func build(root *rootOpts) *buildOpts {
	return &buildOpts{
		rootOpts: root,
		use:      "build [infile] [outfile]",
		short:    "Combine and complete envelope data",
	}
}

func envelop() *buildOpts {
	return &buildOpts{
		envelop: true,
		use:     "envelop [infile] [outfile]",
		short:   "Prepare a document and insert into a new envelope",
	}
}

func (b *buildOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(2),
		RunE:  b.runE,
		Use:   b.use,
		Short: b.short,
	}

	f := cmd.Flags()
	f.StringToStringVar(&b.set, "set", nil, "set value from the command line")
	f.StringToStringVar(&b.setFiles, "set-file", nil, "set value from the specified YAML or JSON file")
	f.StringToStringVar(&b.setStrings, "set-string", nil, "set STRING value from the command line")
	f.StringVarP(&b.template, "template", "T", "", "Template YAML/JSON file into which data is merged")
	f.StringVarP(&b.privateKeyFile, "key", "k", "~/.gobl/id_es256.jwk", "Private key file for signing")
	f.StringVarP(&b.docType, "type", "t", "", "Specify the document type")

	return cmd
}

func (b *buildOpts) outputFilename(args []string) string {
	if b.inPlace {
		return inputFilename(args)
	}
	if len(args) >= 2 && args[1] != "-" {
		return args[1]
	}
	return ""
}

func cmdContext(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func (b *buildOpts) runE(cmd *cobra.Command, args []string) error {
	ctx := cmdContext(cmd)

	var template io.Reader
	if b.template != "" {
		f, err := os.Open(b.template)
		if err != nil {
			return err
		}
		defer f.Close() // nolint:errcheck
		template = f
	}

	input, err := openInput(cmd, args)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if outFile := b.outputFilename(args); outFile != "" {
		flags := os.O_CREATE | os.O_WRONLY
		if !b.overwriteOutputFile && !b.inPlace {
			flags |= os.O_EXCL
		}
		f, err := os.OpenFile(outFile, flags, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close() // nolint:errcheck
		out = f
	} else if b.inPlace {
		return errors.New("cannot overwrite STDIN")
	}
	defer input.Close() // nolint:errcheck

	pkFilename, err := expandHome(b.privateKeyFile)
	if err != nil {
		return err
	}
	keyFile, err := os.Open(pkFilename)
	if err != nil {
		return err
	}
	defer keyFile.Close() // nolint:errcheck

	key := new(dsig.PrivateKey)
	if err = json.NewDecoder(keyFile).Decode(key); err != nil {
		return err
	}

	opts := &internal.BuildOptions{
		Template:   template,
		Data:       input,
		SetFile:    b.setFiles,
		SetYAML:    b.set,
		SetString:  b.setStrings,
		PrivateKey: key,
		DocType:    b.docType,
	}
	var env *gobl.Envelope

	// We're performing the envelop check here to save extra code
	if b.envelop {
		env, err = internal.Envelop(ctx, opts)
	} else {
		env, err = internal.Build(ctx, opts)
	}
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	if b.indent {
		enc.SetIndent("", "\t") // Removing JSON formatting by default
	}
	return enc.Encode(env)
}
