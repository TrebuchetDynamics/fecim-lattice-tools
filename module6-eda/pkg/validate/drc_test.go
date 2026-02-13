package validate

import (
	"testing"
)

func TestValidateLEFDRC_ValidLEFPasses(t *testing.T) {
	lef := `VERSION 5.8 ;
LAYER met1
  TYPE ROUTING ;
  WIDTH 0.14 ;
END met1
MACRO ok_cell
  SIZE 1.0 BY 1.0 ;
  PIN A
    PORT
      LAYER met1 ;
      RECT 0.00 0.00 0.14 0.30 ;
    END
  END A
  PIN B
    PORT
      LAYER met1 ;
      RECT 0.40 0.00 0.54 0.30 ;
    END
  END B
END ok_cell
END LIBRARY`

	if err := ValidateLEFDRC(lef, DefaultSKY130DRCRules()); err != nil {
		t.Fatalf("expected valid LEF to pass DRC, got: %v", err)
	}
}

func TestValidateLEFDRC_BadLEFFails(t *testing.T) {
	lef := `VERSION 5.8 ;
LAYER met1
  TYPE ROUTING ;
  WIDTH 0.10 ;
END met1
MACRO bad_cell
  SIZE 1.0 BY 1.0 ;
  PIN A
    PORT
      LAYER met1 ;
      RECT 0.00 0.00 0.10 0.20 ;
    END
  END A
  PIN B
    PORT
      LAYER met1 ;
      RECT 0.18 0.00 0.28 0.20 ;
    END
  END B
END bad_cell
END LIBRARY`

	if err := ValidateLEFDRC(lef, DefaultSKY130DRCRules()); err == nil {
		t.Fatal("expected intentionally bad LEF to fail DRC")
	}
}

func TestValidateLEFDRC_ViaEnclosureFails(t *testing.T) {
	lef := `VERSION 5.8 ;
MACRO via_bad
  SIZE 1.0 BY 1.0 ;
  OBS
    LAYER met1 ;
    RECT 0.20 0.20 0.40 0.40 ;
    LAYER via1 ;
    RECT 0.18 0.18 0.42 0.42 ;
  END
END via_bad
END LIBRARY`

	if err := ValidateLEFDRC(lef, DefaultSKY130DRCRules()); err == nil {
		t.Fatal("expected via enclosure violation")
	}
}
