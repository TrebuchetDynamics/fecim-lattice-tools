// pkg/export/liberty.go
package export

import (
	"fmt"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/config"
)

// GenerateLiberty generates a Liberty (.lib) timing file for the FeCIM bitcell
// Liberty files provide timing, power, and electrical characteristics for synthesis tools
func GenerateLiberty(cfg config.CellConfig) string {
	area := cfg.Width * cfg.Height
	
	return fmt.Sprintf(`library(fecim_cells) {
  technology (cmos) ;
  delay_model : table_lookup ;
  
  time_unit : "1ns" ;
  voltage_unit : "1V" ;
  current_unit : "1mA" ;
  capacitive_load_unit (1, pf) ;
  leakage_power_unit : "1nW" ;
  
  operating_conditions(typical) {
    process : 1.0 ;
    temperature : 25 ;
    voltage : 1.8 ;
  }
  default_operating_conditions : typical ;
  
  cell(%s) {
    area : %.4f ;
    cell_leakage_power : %.4f ;
    
    pin(WL) {
      direction : input ;
      capacitance : %.4f ;
    }
    
    pin(BL) {
      direction : output ;
      function : "WL" ;
      
      timing() {
        related_pin : "WL" ;
        timing_sense : positive_unate ;
        
        cell_rise(scalar) {
          values("%.3f") ;
        }
        cell_fall(scalar) {
          values("%.3f") ;
        }
        rise_transition(scalar) {
          values("0.050") ;
        }
        fall_transition(scalar) {
          values("0.050") ;
        }
      }
    }
    
    pin(VPWR) {
      direction : inout ;
      pg_type : primary_power ;
    }
    
    pin(VGND) {
      direction : inout ;
      pg_type : primary_ground ;
    }
  }
}
`, cfg.Name, area, cfg.LeakagePower, cfg.InputCap, cfg.RiseTime, cfg.FallTime)
}
