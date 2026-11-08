package models

import (
    "errors"
    "strings"
    "testing"
)
func TestValidateResourceInput(t *testing.T) {
    zero, negative:=0,-1
    base:=CreateResourceInput{Name:"Desk 101",Type:ResourceDesk,Capacity:1,IsActive:true}
    if err:=ValidateResourceInput(base);err!=nil {t.Fatal(err)}
    for _,tc:=range []struct{name string; input CreateResourceInput}{
        {"empty name",func()CreateResourceInput{v:=base;v.Name="";return v}()},
        {"spaces name",func()CreateResourceInput{v:=base;v.Name="  ";return v}()},
        {"padded name",func()CreateResourceInput{v:=base;v.Name=" Desk";return v}()},
        {"control characters",func()CreateResourceInput{v:=base;v.Name="Desk\n1";return v}()},
        {"long name",func()CreateResourceInput{v:=base;v.Name=strings.Repeat("x",151);return v}()},
        {"negative floor",func()CreateResourceInput{v:=base;v.Floor=&negative;return v}()},
        {"zero capacity",func()CreateResourceInput{v:=base;v.Capacity=0;return v}()},
        {"negative capacity",func()CreateResourceInput{v:=base;v.Capacity=-1;return v}()},
        {"invalid kind",func()CreateResourceInput{v:=base;v.Type="INVALID";return v}()},
    }{
        t.Run(tc.name,func(t *testing.T){
            if err:=ValidateResourceInput(tc.input);!errors.Is(err,ErrInvalidResource){
                t.Fatalf("expected ErrInvalidResource, got %v",err)
            }
        })
    }
    okay:=base
    okay.Floor=&zero
    if err:=ValidateResourceInput(okay);err!=nil{t.Fatal(err)}
}
