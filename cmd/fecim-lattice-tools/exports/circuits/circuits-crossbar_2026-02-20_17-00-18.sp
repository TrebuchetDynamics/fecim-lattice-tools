* Module4 Circuits crossbar behavioral deck
.param RWL=0.184 RBL=1.088
.param RWLDRV=0.184 RBLDRV=1.088


* ===== Peripheral behavioral subcircuits =====
.subckt DAC5 vin vout vss
Rdac vout vin 1k
Edac vout vss vin vss 1
.ends DAC5

.subckt SAMPLE_HOLD vin vout vclk vss
Rsw vin n_hold 1000
Chold n_hold vss 1e-12
Rleak n_hold vss 5e+09
Ebuf vout vss n_hold vss 1
.ends SAMPLE_HOLD

.subckt TIA_BASIC iin vout vss
Rtia vout iin 10000
Voff vout n_off 0.005
Rclamp n_off vss 1e12
.ends TIA_BASIC

.subckt ADC5 vin vcode vss
Eadc vcode vss vin vss 15
Radc vcode vss 1e9
.ends ADC5

.subckt VREG_BASIC vin vout vss
Ereg nreg vss vin vss 1
Rdrop nreg vout 0.5
Rpsrr vout vss 1
.ends VREG_BASIC
* =============================================

* Regulated supply for peripherals
VDD_RAW vdd_raw 0 1.8
XREG vdd_raw vdd_periph 0 VREG_BASIC

VWL_SRC_0 wl_src_0 0 0
XDAC_WL_0 wl_src_0 wl_drv_0 0 DAC5
RWL_DRV_0 wl_drv_0 wl_0_0 {RWLDRV}
VWL_SRC_1 wl_src_1 0 0
XDAC_WL_1 wl_src_1 wl_drv_1 0 DAC5
RWL_DRV_1 wl_drv_1 wl_1_0 {RWLDRV}
VWL_SRC_2 wl_src_2 0 0
XDAC_WL_2 wl_src_2 wl_drv_2 0 DAC5
RWL_DRV_2 wl_drv_2 wl_2_0 {RWLDRV}
VWL_SRC_3 wl_src_3 0 0
XDAC_WL_3 wl_src_3 wl_drv_3 0 DAC5
RWL_DRV_3 wl_drv_3 wl_3_0 {RWLDRV}
VWL_SRC_4 wl_src_4 0 0
XDAC_WL_4 wl_src_4 wl_drv_4 0 DAC5
RWL_DRV_4 wl_drv_4 wl_4_0 {RWLDRV}
VWL_SRC_5 wl_src_5 0 0
XDAC_WL_5 wl_src_5 wl_drv_5 0 DAC5
RWL_DRV_5 wl_drv_5 wl_5_0 {RWLDRV}
VWL_SRC_6 wl_src_6 0 0
XDAC_WL_6 wl_src_6 wl_drv_6 0 DAC5
RWL_DRV_6 wl_drv_6 wl_6_0 {RWLDRV}
VWL_SRC_7 wl_src_7 0 0
XDAC_WL_7 wl_src_7 wl_drv_7 0 DAC5
RWL_DRV_7 wl_drv_7 wl_7_0 {RWLDRV}

VBL_SRC_0 bl_src_0 0 0.35
RBL_DRV_0 bl_src_0 bl_0_0 {RBLDRV}
VBL_SRC_1 bl_src_1 0 0
RBL_DRV_1 bl_src_1 bl_0_1 {RBLDRV}
VBL_SRC_2 bl_src_2 0 0
RBL_DRV_2 bl_src_2 bl_0_2 {RBLDRV}
VBL_SRC_3 bl_src_3 0 0
RBL_DRV_3 bl_src_3 bl_0_3 {RBLDRV}
VBL_SRC_4 bl_src_4 0 0
RBL_DRV_4 bl_src_4 bl_0_4 {RBLDRV}
VBL_SRC_5 bl_src_5 0 0
RBL_DRV_5 bl_src_5 bl_0_5 {RBLDRV}
VBL_SRC_6 bl_src_6 0 0
RBL_DRV_6 bl_src_6 bl_0_6 {RBLDRV}
VBL_SRC_7 bl_src_7 0 0
RBL_DRV_7 bl_src_7 bl_0_7 {RBLDRV}

* WL wire resistances
RWL_0_0 wl_0_0 wl_0_1 {RWL}
RWL_0_1 wl_0_1 wl_0_2 {RWL}
RWL_0_2 wl_0_2 wl_0_3 {RWL}
RWL_0_3 wl_0_3 wl_0_4 {RWL}
RWL_0_4 wl_0_4 wl_0_5 {RWL}
RWL_0_5 wl_0_5 wl_0_6 {RWL}
RWL_0_6 wl_0_6 wl_0_7 {RWL}
RWL_1_0 wl_1_0 wl_1_1 {RWL}
RWL_1_1 wl_1_1 wl_1_2 {RWL}
RWL_1_2 wl_1_2 wl_1_3 {RWL}
RWL_1_3 wl_1_3 wl_1_4 {RWL}
RWL_1_4 wl_1_4 wl_1_5 {RWL}
RWL_1_5 wl_1_5 wl_1_6 {RWL}
RWL_1_6 wl_1_6 wl_1_7 {RWL}
RWL_2_0 wl_2_0 wl_2_1 {RWL}
RWL_2_1 wl_2_1 wl_2_2 {RWL}
RWL_2_2 wl_2_2 wl_2_3 {RWL}
RWL_2_3 wl_2_3 wl_2_4 {RWL}
RWL_2_4 wl_2_4 wl_2_5 {RWL}
RWL_2_5 wl_2_5 wl_2_6 {RWL}
RWL_2_6 wl_2_6 wl_2_7 {RWL}
RWL_3_0 wl_3_0 wl_3_1 {RWL}
RWL_3_1 wl_3_1 wl_3_2 {RWL}
RWL_3_2 wl_3_2 wl_3_3 {RWL}
RWL_3_3 wl_3_3 wl_3_4 {RWL}
RWL_3_4 wl_3_4 wl_3_5 {RWL}
RWL_3_5 wl_3_5 wl_3_6 {RWL}
RWL_3_6 wl_3_6 wl_3_7 {RWL}
RWL_4_0 wl_4_0 wl_4_1 {RWL}
RWL_4_1 wl_4_1 wl_4_2 {RWL}
RWL_4_2 wl_4_2 wl_4_3 {RWL}
RWL_4_3 wl_4_3 wl_4_4 {RWL}
RWL_4_4 wl_4_4 wl_4_5 {RWL}
RWL_4_5 wl_4_5 wl_4_6 {RWL}
RWL_4_6 wl_4_6 wl_4_7 {RWL}
RWL_5_0 wl_5_0 wl_5_1 {RWL}
RWL_5_1 wl_5_1 wl_5_2 {RWL}
RWL_5_2 wl_5_2 wl_5_3 {RWL}
RWL_5_3 wl_5_3 wl_5_4 {RWL}
RWL_5_4 wl_5_4 wl_5_5 {RWL}
RWL_5_5 wl_5_5 wl_5_6 {RWL}
RWL_5_6 wl_5_6 wl_5_7 {RWL}
RWL_6_0 wl_6_0 wl_6_1 {RWL}
RWL_6_1 wl_6_1 wl_6_2 {RWL}
RWL_6_2 wl_6_2 wl_6_3 {RWL}
RWL_6_3 wl_6_3 wl_6_4 {RWL}
RWL_6_4 wl_6_4 wl_6_5 {RWL}
RWL_6_5 wl_6_5 wl_6_6 {RWL}
RWL_6_6 wl_6_6 wl_6_7 {RWL}
RWL_7_0 wl_7_0 wl_7_1 {RWL}
RWL_7_1 wl_7_1 wl_7_2 {RWL}
RWL_7_2 wl_7_2 wl_7_3 {RWL}
RWL_7_3 wl_7_3 wl_7_4 {RWL}
RWL_7_4 wl_7_4 wl_7_5 {RWL}
RWL_7_5 wl_7_5 wl_7_6 {RWL}
RWL_7_6 wl_7_6 wl_7_7 {RWL}

* BL wire resistances
RBL_0_0 bl_0_0 bl_1_0 {RBL}
RBL_1_0 bl_1_0 bl_2_0 {RBL}
RBL_2_0 bl_2_0 bl_3_0 {RBL}
RBL_3_0 bl_3_0 bl_4_0 {RBL}
RBL_4_0 bl_4_0 bl_5_0 {RBL}
RBL_5_0 bl_5_0 bl_6_0 {RBL}
RBL_6_0 bl_6_0 bl_7_0 {RBL}
RBL_0_1 bl_0_1 bl_1_1 {RBL}
RBL_1_1 bl_1_1 bl_2_1 {RBL}
RBL_2_1 bl_2_1 bl_3_1 {RBL}
RBL_3_1 bl_3_1 bl_4_1 {RBL}
RBL_4_1 bl_4_1 bl_5_1 {RBL}
RBL_5_1 bl_5_1 bl_6_1 {RBL}
RBL_6_1 bl_6_1 bl_7_1 {RBL}
RBL_0_2 bl_0_2 bl_1_2 {RBL}
RBL_1_2 bl_1_2 bl_2_2 {RBL}
RBL_2_2 bl_2_2 bl_3_2 {RBL}
RBL_3_2 bl_3_2 bl_4_2 {RBL}
RBL_4_2 bl_4_2 bl_5_2 {RBL}
RBL_5_2 bl_5_2 bl_6_2 {RBL}
RBL_6_2 bl_6_2 bl_7_2 {RBL}
RBL_0_3 bl_0_3 bl_1_3 {RBL}
RBL_1_3 bl_1_3 bl_2_3 {RBL}
RBL_2_3 bl_2_3 bl_3_3 {RBL}
RBL_3_3 bl_3_3 bl_4_3 {RBL}
RBL_4_3 bl_4_3 bl_5_3 {RBL}
RBL_5_3 bl_5_3 bl_6_3 {RBL}
RBL_6_3 bl_6_3 bl_7_3 {RBL}
RBL_0_4 bl_0_4 bl_1_4 {RBL}
RBL_1_4 bl_1_4 bl_2_4 {RBL}
RBL_2_4 bl_2_4 bl_3_4 {RBL}
RBL_3_4 bl_3_4 bl_4_4 {RBL}
RBL_4_4 bl_4_4 bl_5_4 {RBL}
RBL_5_4 bl_5_4 bl_6_4 {RBL}
RBL_6_4 bl_6_4 bl_7_4 {RBL}
RBL_0_5 bl_0_5 bl_1_5 {RBL}
RBL_1_5 bl_1_5 bl_2_5 {RBL}
RBL_2_5 bl_2_5 bl_3_5 {RBL}
RBL_3_5 bl_3_5 bl_4_5 {RBL}
RBL_4_5 bl_4_5 bl_5_5 {RBL}
RBL_5_5 bl_5_5 bl_6_5 {RBL}
RBL_6_5 bl_6_5 bl_7_5 {RBL}
RBL_0_6 bl_0_6 bl_1_6 {RBL}
RBL_1_6 bl_1_6 bl_2_6 {RBL}
RBL_2_6 bl_2_6 bl_3_6 {RBL}
RBL_3_6 bl_3_6 bl_4_6 {RBL}
RBL_4_6 bl_4_6 bl_5_6 {RBL}
RBL_5_6 bl_5_6 bl_6_6 {RBL}
RBL_6_6 bl_6_6 bl_7_6 {RBL}
RBL_0_7 bl_0_7 bl_1_7 {RBL}
RBL_1_7 bl_1_7 bl_2_7 {RBL}
RBL_2_7 bl_2_7 bl_3_7 {RBL}
RBL_3_7 bl_3_7 bl_4_7 {RBL}
RBL_4_7 bl_4_7 bl_5_7 {RBL}
RBL_5_7 bl_5_7 bl_6_7 {RBL}
RBL_6_7 bl_6_7 bl_7_7 {RBL}

* Memory cell conductances
RCELL_0_0 wl_0_0 bl_0_0 92367.0857
RCELL_0_1 wl_0_1 bl_0_1 92367.0857
RCELL_0_2 wl_0_2 bl_0_2 92367.0857
RCELL_0_3 wl_0_3 bl_0_3 92367.0857
RCELL_0_4 wl_0_4 bl_0_4 92367.0857
RCELL_0_5 wl_0_5 bl_0_5 92367.0857
RCELL_0_6 wl_0_6 bl_0_6 92367.0857
RCELL_0_7 wl_0_7 bl_0_7 92367.0857
RCELL_1_0 wl_1_0 bl_1_0 92367.0857
RCELL_1_1 wl_1_1 bl_1_1 92367.0857
RCELL_1_2 wl_1_2 bl_1_2 92367.0857
RCELL_1_3 wl_1_3 bl_1_3 92367.0857
RCELL_1_4 wl_1_4 bl_1_4 92367.0857
RCELL_1_5 wl_1_5 bl_1_5 92367.0857
RCELL_1_6 wl_1_6 bl_1_6 92367.0857
RCELL_1_7 wl_1_7 bl_1_7 92367.0857
RCELL_2_0 wl_2_0 bl_2_0 92367.0857
RCELL_2_1 wl_2_1 bl_2_1 92367.0857
RCELL_2_2 wl_2_2 bl_2_2 92367.0857
RCELL_2_3 wl_2_3 bl_2_3 92367.0857
RCELL_2_4 wl_2_4 bl_2_4 92367.0857
RCELL_2_5 wl_2_5 bl_2_5 92367.0857
RCELL_2_6 wl_2_6 bl_2_6 92367.0857
RCELL_2_7 wl_2_7 bl_2_7 92367.0857
RCELL_3_0 wl_3_0 bl_3_0 92367.0857
RCELL_3_1 wl_3_1 bl_3_1 92367.0857
RCELL_3_2 wl_3_2 bl_3_2 92367.0857
RCELL_3_3 wl_3_3 bl_3_3 92367.0857
RCELL_3_4 wl_3_4 bl_3_4 92367.0857
RCELL_3_5 wl_3_5 bl_3_5 92367.0857
RCELL_3_6 wl_3_6 bl_3_6 92367.0857
RCELL_3_7 wl_3_7 bl_3_7 92367.0857
RCELL_4_0 wl_4_0 bl_4_0 92367.0857
RCELL_4_1 wl_4_1 bl_4_1 92367.0857
RCELL_4_2 wl_4_2 bl_4_2 92367.0857
RCELL_4_3 wl_4_3 bl_4_3 92367.0857
RCELL_4_4 wl_4_4 bl_4_4 92367.0857
RCELL_4_5 wl_4_5 bl_4_5 92367.0857
RCELL_4_6 wl_4_6 bl_4_6 92367.0857
RCELL_4_7 wl_4_7 bl_4_7 92367.0857
RCELL_5_0 wl_5_0 bl_5_0 92367.0857
RCELL_5_1 wl_5_1 bl_5_1 92367.0857
RCELL_5_2 wl_5_2 bl_5_2 92367.0857
RCELL_5_3 wl_5_3 bl_5_3 92367.0857
RCELL_5_4 wl_5_4 bl_5_4 92367.0857
RCELL_5_5 wl_5_5 bl_5_5 92367.0857
RCELL_5_6 wl_5_6 bl_5_6 92367.0857
RCELL_5_7 wl_5_7 bl_5_7 92367.0857
RCELL_6_0 wl_6_0 bl_6_0 92367.0857
RCELL_6_1 wl_6_1 bl_6_1 92367.0857
RCELL_6_2 wl_6_2 bl_6_2 92367.0857
RCELL_6_3 wl_6_3 bl_6_3 92367.0857
RCELL_6_4 wl_6_4 bl_6_4 92367.0857
RCELL_6_5 wl_6_5 bl_6_5 92367.0857
RCELL_6_6 wl_6_6 bl_6_6 92367.0857
RCELL_6_7 wl_6_7 bl_6_7 92367.0857
RCELL_7_0 wl_7_0 bl_7_0 92367.0857
RCELL_7_1 wl_7_1 bl_7_1 92367.0857
RCELL_7_2 wl_7_2 bl_7_2 92367.0857
RCELL_7_3 wl_7_3 bl_7_3 92367.0857
RCELL_7_4 wl_7_4 bl_7_4 92367.0857
RCELL_7_5 wl_7_5 bl_7_5 92367.0857
RCELL_7_6 wl_7_6 bl_7_6 92367.0857
RCELL_7_7 wl_7_7 bl_7_7 92367.0857

* Readout peripherals per BL
XSH_0 bl_7_0 bl_sh_0 sh_clk 0 SAMPLE_HOLD
XTIA_0 bl_sh_0 vout_0 0 TIA_BASIC
XADC_0 vout_0 code_0 0 ADC5
XSH_1 bl_7_1 bl_sh_1 sh_clk 0 SAMPLE_HOLD
XTIA_1 bl_sh_1 vout_1 0 TIA_BASIC
XADC_1 vout_1 code_1 0 ADC5
XSH_2 bl_7_2 bl_sh_2 sh_clk 0 SAMPLE_HOLD
XTIA_2 bl_sh_2 vout_2 0 TIA_BASIC
XADC_2 vout_2 code_2 0 ADC5
XSH_3 bl_7_3 bl_sh_3 sh_clk 0 SAMPLE_HOLD
XTIA_3 bl_sh_3 vout_3 0 TIA_BASIC
XADC_3 vout_3 code_3 0 ADC5
XSH_4 bl_7_4 bl_sh_4 sh_clk 0 SAMPLE_HOLD
XTIA_4 bl_sh_4 vout_4 0 TIA_BASIC
XADC_4 vout_4 code_4 0 ADC5
XSH_5 bl_7_5 bl_sh_5 sh_clk 0 SAMPLE_HOLD
XTIA_5 bl_sh_5 vout_5 0 TIA_BASIC
XADC_5 vout_5 code_5 0 ADC5
XSH_6 bl_7_6 bl_sh_6 sh_clk 0 SAMPLE_HOLD
XTIA_6 bl_sh_6 vout_6 0 TIA_BASIC
XADC_6 vout_6 code_6 0 ADC5
XSH_7 bl_7_7 bl_sh_7 sh_clk 0 SAMPLE_HOLD
XTIA_7 bl_sh_7 vout_7 0 TIA_BASIC
XADC_7 vout_7 code_7 0 ADC5

.control
op
print all
.endc

.end
