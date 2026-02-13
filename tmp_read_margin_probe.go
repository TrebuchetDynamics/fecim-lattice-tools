package main

import (
  "fmt"
  "fecim-lattice-tools/module4-circuits/pkg/arraysim"
  "fecim-lattice-tools/shared/physics"
)

func main(){
 mat:=physics.DefaultHZO()
 g14:=mat.DiscreteLevel(14,30)
 g15:=mat.DiscreteLevel(15,30)
 g16:=mat.DiscreteLevel(16,30)
 fmt.Println("G",g14,g15,g16)
 sense:=arraysim.SenseChain{TIA:arraysim.TIAConfig{Rf:20e3,Vref:0.2,Vmin:0,Vmax:1.2},ADC:arraysim.ADCConfig{Bits:8,Vmin:0,Vmax:1.2}}
 sizes:=[]int{8,16,32,64,128}
 rons:=[]float64{0,100,500,1e3,5e3,1e4}
 levels:=[]float64{g14,g15,g16}
 for _,n:=range sizes{
  tr:=n/2; tc:=n/2
  fmt.Println("size",n)
  for _,ron:= range rons{
    codes:=make([]int,3)
    for li,g:= range levels {
      cond:=make([][]float64,n)
      for r:=0;r<n;r++{cond[r]=make([]float64,n); for c:=0;c<n;c++{cond[r][c]=g15}}
      cond[tr][tc]=g
      wl:=make([]float64,n); bl:=make([]float64,n)
      wl[tr]=0.02
      readMask:=make([][]bool,n)
      for r:=0;r<n;r++{readMask[r]=make([]bool,n)}
      readMask[tr][tc]=true
      selector:=arraysim.SelectorDeviceParams{}
      if ron>0{selector=arraysim.SelectorDeviceParams{Enabled:true,OnConductance:1.0/ron,OffConductance:1e-12}}
      params:=arraysim.SolveParams{WLVoltages:wl,BLVoltages:bl,Conductance:cond,SelectorMode:arraysim.SelectorRead,ReadMask:readMask,Selector:selector}
      res,err:=arraysim.NewTierBSolver().Solve(params)
      if err!=nil{panic(err)}
      sr:=sense.ConvertCurrent(res.RowCurrents[tr])
      codes[li]=sr.Code
    }
    d1:=codes[1]-codes[0]; if d1<0{d1=-d1}
    d2:=codes[2]-codes[1]; if d2<0{d2=-d2}
    m:=d1; if d2<m{m=d2}
    fmt.Printf(" ron %.0f codes %v m %d\n",ron,codes,m)
  }
 }
}
