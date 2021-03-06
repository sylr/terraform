package arguments

import (
	"flag"
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/tfdiags"
)

// DefaultParallelism is the limit Terraform places on total parallel
// operations as it walks the dependency graph.
const DefaultParallelism = 10

// State describes arguments which are used to define how Terraform interacts
// with state.
type State struct {
	// Lock controls whether or not the state manager is used to lock state
	// during operations.
	Lock bool

	// LockTimeout allows setting a time limit on acquiring the state lock.
	// The default is 0, meaning no limit.
	LockTimeout time.Duration

	// StatePath specifies a non-default location for the state file. The
	// default value is blank, which is interpeted as "terraform.tfstate".
	StatePath string

	// StateOutPath specifies a different path to write the final state file.
	// The default value is blank, which results in state being written back to
	// StatePath.
	StateOutPath string

	// BackupPath specifies the path where a backup copy of the state file will
	// be stored before the new state is written. The default value is blank,
	// which is interpreted as StateOutPath +
	// ".backup".
	BackupPath string
}

// Operation describes arguments which are used to configure how a Terraform
// operation such as a plan or apply executes.
type Operation struct {
	// Parallelism is the limit Terraform places on total parallel operations
	// as it walks the dependency graph.
	Parallelism int

	// Refresh controls whether or not the operation should refresh existing
	// state before proceeding. Default is true.
	Refresh bool

	// Targets allow limiting an operation to a set of resource addresses and
	// their dependencies.
	Targets []addrs.Targetable

	targetsRaw []string
}

// Parse must be called on Operation after initial flag parse. This processes
// the raw target flags into addrs.Targetable values, returning diagnostics if
// invalid.
func (o *Operation) Parse() tfdiags.Diagnostics {
	var diags tfdiags.Diagnostics

	o.Targets = nil

	for _, tr := range o.targetsRaw {
		traversal, syntaxDiags := hclsyntax.ParseTraversalAbs([]byte(tr), "", hcl.Pos{Line: 1, Column: 1})
		if syntaxDiags.HasErrors() {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				fmt.Sprintf("Invalid target %q", tr),
				syntaxDiags[0].Detail,
			))
			continue
		}

		target, targetDiags := addrs.ParseTarget(traversal)
		if targetDiags.HasErrors() {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				fmt.Sprintf("Invalid target %q", tr),
				targetDiags[0].Description().Detail,
			))
			continue
		}

		o.Targets = append(o.Targets, target.Subject)
	}

	return diags
}

// Vars describes arguments which specify non-default variable values. This
// interfce is unfortunately obscure, because the order of the CLI arguments
// determines the final value of the gathered variables. In future it might be
// desirable for the arguments package to handle the gathering of variables
// directly, returning a map of variable values.
type Vars struct {
	vars     *flagNameValueSlice
	varFiles *flagNameValueSlice
}

func (v *Vars) All() []FlagNameValue {
	if v.vars == nil {
		return nil
	}
	return v.vars.AllItems()
}

func (v *Vars) Empty() bool {
	if v.vars == nil {
		return true
	}
	return v.vars.Empty()
}

// extendedFlagSet creates a FlagSet with common backend, operation, and vars
// flags used in many commands. Target structs for each subset of flags must be
// provided in order to support those flags.
func extendedFlagSet(name string, state *State, operation *Operation, vars *Vars) *flag.FlagSet {
	f := defaultFlagSet(name)

	if state == nil && operation == nil && vars == nil {
		panic("use defaultFlagSet")
	}

	if state != nil {
		f.BoolVar(&state.Lock, "lock", true, "lock")
		f.DurationVar(&state.LockTimeout, "lock-timeout", 0, "lock-timeout")
		f.StringVar(&state.StatePath, "state", "", "state-path")
		f.StringVar(&state.StateOutPath, "state-out", "", "state-path")
		f.StringVar(&state.BackupPath, "backup", "", "backup-path")
	}

	if operation != nil {
		f.IntVar(&operation.Parallelism, "parallelism", DefaultParallelism, "parallelism")
		f.BoolVar(&operation.Refresh, "refresh", true, "refresh")
		f.Var((*flagStringSlice)(&operation.targetsRaw), "target", "target")
	}

	// Gather all -var and -var-file arguments into one heterogenous structure
	// to preserve the overall order.
	if vars != nil {
		varsFlags := newFlagNameValueSlice("-var")
		varFilesFlags := varsFlags.Alias("-var-file")
		vars.vars = &varsFlags
		vars.varFiles = &varFilesFlags
		f.Var(vars.vars, "var", "var")
		f.Var(vars.varFiles, "var-file", "var-file")
	}

	return f
}
