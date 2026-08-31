package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-improvement-selector/selector"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gooo-selector <compile|generate|measure|select|report>")
	}
	switch args[0] {
	case "compile":
		return compileCommand(args[1:])
	case "generate":
		return generateCommand(args[1:])
	case "measure":
		return measureCommand(args[1:])
	case "select":
		return selectCommand(args[1:])
	case "report":
		return reportCommand(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func compileCommand(args []string) error {
	fs := flag.NewFlagSet("compile", flag.ContinueOnError)
	sourcePath := fs.String("source", "", "Gooo source")
	contractPath := fs.String("contract", "", "fixed denominator")
	outPath := fs.String("out", "", "semantic IR output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *sourcePath == "" || *contractPath == "" || *outPath == "" {
		return fmt.Errorf("compile requires -source, -contract, and -out")
	}
	sourceDigest, source, err := selector.DigestFile(*sourcePath)
	if err != nil {
		return err
	}
	contractDigest, _, err := selector.DigestFile(*contractPath)
	if err != nil {
		return err
	}
	var contract selector.Contract
	if err := selector.ReadJSON(*contractPath, &contract); err != nil {
		return err
	}
	ir, err := selector.Compile(string(source), contract, contractDigest)
	if err != nil {
		return err
	}
	if ir.SourceDigest != sourceDigest {
		return fmt.Errorf("source digest changed during compilation")
	}
	return selector.WriteJSON(*outPath, ir)
}

func generateCommand(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	irPath := fs.String("ir", "", "semantic IR")
	scenarioPath := fs.String("scenario", "", "scenario input")
	inputPath := fs.String("input", "", "subject input")
	contractPath := fs.String("contract", "", "fixed denominator")
	outDir := fs.String("out-dir", "", "caller-owned candidate output directory")
	manifestPath := fs.String("manifest", "", "candidate manifest output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *irPath == "" || *scenarioPath == "" || *inputPath == "" || *contractPath == "" || *outDir == "" || *manifestPath == "" {
		return fmt.Errorf("generate requires -ir, -scenario, -input, -contract, -out-dir, and -manifest")
	}
	var ir selector.SemanticIR
	if err := selector.ReadJSON(*irPath, &ir); err != nil {
		return err
	}
	var scenario selector.Scenario
	if err := selector.ReadJSON(*scenarioPath, &scenario); err != nil {
		return err
	}
	inputDigest, _, err := selector.DigestFile(*inputPath)
	if err != nil {
		return err
	}
	contractDigest, _, err := selector.DigestFile(*contractPath)
	if err != nil {
		return err
	}
	irDigest, _, err := selector.DigestFile(*irPath)
	if err != nil {
		return err
	}
	manifest, err := selector.GenerateCandidates(ir, scenario, inputDigest, contractDigest, irDigest, *outDir)
	if err != nil {
		return err
	}
	return selector.WriteJSON(*manifestPath, manifest)
}

func measureCommand(args []string) error {
	fs := flag.NewFlagSet("measure", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "candidate manifest")
	scenarioPath := fs.String("scenario", "", "scenario input")
	inputPath := fs.String("input", "", "subject input")
	contractPath := fs.String("contract", "", "fixed denominator")
	outPath := fs.String("out", "", "measurement receipt output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *scenarioPath == "" || *inputPath == "" || *contractPath == "" || *outPath == "" {
		return fmt.Errorf("measure requires -manifest, -scenario, -input, -contract, and -out")
	}
	var manifest selector.CandidateManifest
	if err := selector.ReadJSON(*manifestPath, &manifest); err != nil {
		return err
	}
	var scenario selector.Scenario
	if err := selector.ReadJSON(*scenarioPath, &scenario); err != nil {
		return err
	}
	inputDigest, _, err := selector.DigestFile(*inputPath)
	if err != nil {
		return err
	}
	contractDigest, _, err := selector.DigestFile(*contractPath)
	if err != nil {
		return err
	}
	receipt, err := selector.MeasureScenario(scenario, manifest, inputDigest, contractDigest)
	if err != nil {
		return err
	}
	return selector.WriteJSON(*outPath, receipt)
}

func selectCommand(args []string) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	contractPath := fs.String("contract", "", "fixed denominator")
	manifestPath := fs.String("manifest", "", "candidate manifest")
	receiptPath := fs.String("receipt", "", "measurement receipt")
	outPath := fs.String("out", "", "selection output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *contractPath == "" || *manifestPath == "" || *receiptPath == "" || *outPath == "" {
		return fmt.Errorf("select requires -contract, -manifest, -receipt, and -out")
	}
	var contract selector.Contract
	if err := selector.ReadJSON(*contractPath, &contract); err != nil {
		return err
	}
	var manifest selector.CandidateManifest
	if err := selector.ReadJSON(*manifestPath, &manifest); err != nil {
		return err
	}
	var receipt selector.MeasurementReceipt
	if err := selector.ReadJSON(*receiptPath, &receipt); err != nil {
		return err
	}
	selection, err := selector.SelectImprovement(contract, manifest, receipt)
	if err != nil {
		return err
	}
	return selector.WriteJSON(*outPath, selection)
}

func reportCommand(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	irPath := fs.String("ir", "", "semantic IR")
	contractPath := fs.String("contract", "", "fixed denominator")
	summaryPath := fs.String("summary", "", "selection summary")
	runtimePath := fs.String("runtime", "", "runtime metrics")
	outPath := fs.String("out", "", "human report")
	outJSONPath := fs.String("out-json", "", "machine report")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *irPath == "" || *contractPath == "" || *summaryPath == "" || *runtimePath == "" || *outPath == "" || *outJSONPath == "" {
		return fmt.Errorf("report requires -ir, -contract, -summary, -runtime, -out, and -out-json")
	}
	var ir selector.SemanticIR
	if err := selector.ReadJSON(*irPath, &ir); err != nil {
		return err
	}
	var contract selector.Contract
	if err := selector.ReadJSON(*contractPath, &contract); err != nil {
		return err
	}
	var summary struct {
		Schema     string             `json:"schema"`
		Selections []selector.Selection `json:"selections"`
	}
	if err := selector.ReadJSON(*summaryPath, &summary); err != nil {
		return err
	}
	var runtime selector.RuntimeMetrics
	if err := selector.ReadJSON(*runtimePath, &runtime); err != nil {
		return err
	}
	irDigest, _, err := selector.DigestFile(*irPath)
	if err != nil {
		return err
	}
	contractDigest, _, err := selector.DigestFile(*contractPath)
	if err != nil {
		return err
	}
	report, err := selector.BuildReport(ir, contract, irDigest, contractDigest, summary.Selections, runtime)
	if err != nil {
		return err
	}
	if err := selector.WriteJSON(*outJSONPath, report); err != nil {
		return err
	}
	return selector.WriteBytes(*outPath, []byte(selector.RenderMarkdown(report)))
}
