//go:build ignore

package main
import("encoding/json";"os";"fecim-lattice-tools/workbench/moduleapi")
func main(){json.NewEncoder(os.Stdout).Encode(moduleapi.BrowserCatalogue())}
