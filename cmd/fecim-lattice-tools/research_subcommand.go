package main

import researchcmd "fecim-lattice-tools/cmd/fecim-lattice-tools/research"

var researchRunner = researchcmd.RunTool

func runResearchSubcommand(args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return researchRunner(args)
}

func runResearchTool(args []string) error {
	return researchcmd.RunTool(args)
}

func researchRepoRoot(args []string) (string, error) {
	return researchcmd.RepoRoot(args)
}

func repoRootFromResearchArgs(args []string) string {
	return researchcmd.RepoRootFromArgs(args)
}

func normalizeResearchRepoRootArg(args []string, root string) []string {
	return researchcmd.NormalizeRepoRootArg(args, root)
}

func validateResearchRepoRoot(root string) (string, error) {
	return researchcmd.ValidateRepoRoot(root)
}

func findResearchRepoRoot(start string) (string, bool) {
	return researchcmd.FindRepoRoot(start)
}

func researchScriptExists(root string) bool {
	return researchcmd.ScriptExists(root)
}
