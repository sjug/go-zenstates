# go-zenstates

go-zenstates is a p-state overclocking tool for linux. By acccesing the MSR of new AMD Ryzen processors, we can adjust p-states on the fly from the desktop.

All credit goes to [elmor](http://www.overclock.net/u/401414/elmor) from Asus for getting the MSR locations. AMD also has provided more MSR info in their [Ryzen Processor Programming Reference](https://support.amd.com/TechDocs/54945_PPR_Family_17h_Models_00h-0Fh.pdf) (pdf).

## Installation

`go get github.com/sjug/go-zenstates`

## Usage

go-zenstates requires root privileges and the msr module to be loaded. 

To set a 4.0GHz p-state 0 with a 1.4v vcore:
```
sudo modprobe msr
sudo go-zenstates -p 0 -fid 0xA0 -vid 0x18
```

All ID's are hex values with a leading 0x as shown above.

FID is the Frequency ID.  
DID is the Divisor ID.  
VID is the Voltage ID.  

For most purposes there is no reason to change the DID, only the FID and VID to adjust the frequency and voltage.

```
Core Frequency = BCLK*FID/DID
CPU Ratio = 25*FID/(12.5*DID)
Core Voltage = (1.55-0.00625*VID)
```

noko from hardforum has also put together a very easy to use [calculator](https://hardforum.com/threads/ryzen-pstate-overclocking-method-calculation-and-calculator.1928648/#post-1042913631) for these values.

## Help

```
Usage of go-zenstates:           
  -did uint                        
        DID to set (hex)           
  -disable                         
        Disable pstate             
  -enable                          
        Enable pstate              
  -fid uint                        
        FID to set (hex)           
  -l    List all pstates           
  -p int                           
        pstate to set (default -1) 
  -vid uint                        
        VID to set (hex)
```
