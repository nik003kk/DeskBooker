package repository

import (
    "context"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

type transactionStarter interface {
    Begin(context.Context) (pgx.Tx,error)
}

// UpdateResourceWithFeatures updates all fields and equipment atomically.
// A row-level lock prevents concurrent edits from overwriting each other.
func (s *Store) UpdateResourceWithFeatures(ctx context.Context,id string,patch models.ResourcePatch)(models.Resource,[]string,error){
    starter,ok:=s.db.(transactionStarter)
    if !ok {return models.Resource{},nil,errors.New("database does not support transactions")}
    tx,err:=starter.Begin(ctx)
    if err!=nil {return models.Resource{},nil,fmt.Errorf("begin resource update: %w",err)}
    defer func(){_ = tx.Rollback(context.Background())}()

    current,err:=scanResource(tx.QueryRow(ctx,"SELECT "+resourceColumns+" FROM public.resources WHERE id=$1 FOR UPDATE",id))
    if err!=nil{return models.Resource{},nil,fmt.Errorf("load resource: %w",err)}
    next:=models.CreateResourceInput{Name:current.Name,Type:current.Type,Floor:current.Floor,Capacity:current.Capacity,IsActive:current.IsActive}
    if patch.Name!=nil {next.Name=*patch.Name}
    if patch.Type!=nil {next.Type=*patch.Type}
    if patch.FloorSet {next.Floor=patch.Floor}
    if patch.Capacity!=nil {next.Capacity=*patch.Capacity}
    if patch.IsActive!=nil {next.IsActive=*patch.IsActive}
    if err:=models.ValidateResourceInput(next);err!=nil {return models.Resource{},nil,err}

    result,err:=scanResource(tx.QueryRow(ctx,
        "UPDATE public.resources SET name=$2,type=$3,floor=$4,capacity=$5,is_active=$6 WHERE id=$1 RETURNING "+resourceColumns,
        id,next.Name,next.Type,next.Floor,next.Capacity,next.IsActive))
    if err!=nil{return models.Resource{},nil,fmt.Errorf("update resource: %w",err)}
    if patch.FeaturesSet {
        if _,err:=tx.Exec(ctx,"DELETE FROM public.resource_features WHERE resource_id=$1",id);err!=nil {
            return models.Resource{},nil,fmt.Errorf("remove old resource equipment: %w",err)
        }
        if len(patch.Features)>0 {
            if _,err:=tx.Exec(ctx,
                "INSERT INTO public.resource_features (resource_id,feature) SELECT $1::uuid,name FROM unnest($2::text[]) AS name",
                id,patch.Features);err!=nil {
                return models.Resource{},nil,fmt.Errorf("save resource equipment: %w",err)
            }
        }
    }
    stored,err:=New(tx).ListFeatures(ctx,id)
    if err!=nil{return models.Resource{},nil,err}
    features:=make([]string,0,len(stored))
    for _,feature:=range stored {features=append(features,feature.Feature)}
    if err:=tx.Commit(ctx);err!=nil{return models.Resource{},nil,fmt.Errorf("commit resource: %w",err)}
    return result,features,nil
}
