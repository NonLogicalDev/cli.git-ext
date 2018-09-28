package main

var helpTemplate = `{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .Commands}}\
{{if not .Hidden}}\
  {{.FullCommand}}{{if .Default}}*{{end}}{{if .Aliases}} (alias={{.Aliases}}) {{end}}{{template "FormatCommand" .}}
{{.Help|Wrap 4}}
{{if .CmdGroupModel}}\
{{template "FormatCommands" .CmdGroupModel}}
{{end}}\
{{end}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0}}\
{{end}}\

{{end}}\

{{if .Context.SelectedCommand}}\
usage: {{.App.Name}} {{.Context.SelectedCommand}}{{template "FormatUsage" .Context.SelectedCommand}}
{{else}}\
usage: {{.App.Name}}{{template "FormatUsage" .App}}
{{end}}\
{{if .Context.Flags}}\
Flags:
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.Args}}\
Args:
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand}}\
{{if len .Context.SelectedCommand.Commands}}\
Subcommands:
{{template "FormatCommands" .Context.SelectedCommand}}
{{end}}\
{{else if .App.Commands}}\
Commands:
{{template "FormatCommands" .App}}
{{end}}\
`

var helpTemplateLong = `{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .Commands}}\
{{if not .Hidden}}\
  {{.FullCommand}}{{if .Aliases}} (alias={{.Aliases}}) {{end}}{{template "FormatCommand" .}}
{{.Help|Wrap 4}}
{{with .Flags|FlagsToTwoColumns}}{{FormatTwoColumnsWithIndent . 4 2}}{{end}}
{{if .CmdGroupModel}}\
{{template "FormatCommands" .CmdGroupModel}}
{{end}}\
{{end}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0}}\
{{end}}\

{{end}}\

usage: {{.App.Name}}{{template "FormatUsage" .App}}
{{if .Context.Flags}}\
Flags:
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.Args}}\
Args:
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .App.Commands}}\
Commands:
{{template "FormatCommands" .App}}
{{end}}\
`

var helpTemplateManPage = `{{define "FormatFlags"}}\
{{range .Flags}}\
{{if not .Hidden}}\
.TP
\fB{{if .Short}}-{{.Short|Char}}, {{end}}--{{.Name}}{{if not .IsBoolFlag}}={{.FormatPlaceHolder}}{{end}}\\fR
{{.Help}}
{{end}}\
{{end}}\
{{end}}\

{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}{{if .Default}}*{{end}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .Commands}}\
{{if not .Hidden}}\
.SS
\fB{{.FullCommand}}{{if .Aliases}} (alias={{.Aliases}}) {{end}}{{template "FormatCommand" .}}\\fR
.PP
{{.Help}}
{{template "FormatFlags" .}}\
{{if .CmdGroupModel}}\
{{template "FormatCommands" .CmdGroupModel}}
{{end}}\
{{end}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}\\fR
{{end}}\

.TH {{.App.Name}} 1 {{.App.Version}} "{{.App.Author}}"
.SH "NAME"
{{.App.Name}}
.SH "SYNOPSIS"
.TP
\fB{{.App.Name}}{{template "FormatUsage" .App}}
.SH "DESCRIPTION"
{{.App.Help}}
.SH "OPTIONS"
{{template "FormatFlags" .App}}\
{{if .App.Commands}}\
.SH "COMMANDS"
{{template "FormatCommands" .App}}\
{{end}}\
`
