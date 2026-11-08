package repository

import (
  "context"
  "fmt"
  "strings"

  "github.com/nik003kk/DeskBooker/backend/internal/models"
)

// SearchResources returns active resources matching filters.
// A time window excludes overlapping PENDING / CONFIRMED reservations.
// Multiple equipment filters use all-of matching, not any-of.
func (s *Store) SearchResources(ctx context.Context,filter models.ResourceSearch)([]models.Resource,error){
 var kind any
 if filter.Type!=nil {kind=string(*filter.Type)}
 var from,to any
 if filter.StartsAt!=nil {from=*filter.StartsAt}
 if filter.EndsAt!=nil {to=*filter.EndsAt}
 features:=make([]string,0,len(filter.Features))
 for _,name:=range filter.Features {features=append(features,strings.ToLower(name))}
 const query = "SELECT r.id,r.name,r.type,r.floor,r.capacity,r.is_active,r.created_at "+
  "FROM public.resources r "+
  "WHERE r.is_active=TRUE "+
  "AND ($1::text IS NULL OR r.type=$1) "+
  "AND ($2::timestamptz IS NULL OR NOT EXISTS ("+
    "SELECT 1 FROM public.reservations b "+
    "WHERE b.resource_id=r.id AND b.status IN ('PENDING','CONFIRMED') "+
    "AND b.starts_at < $3::timestamptz AND b.ends_at > $2::timestamptz)) "+
  "AND NOT EXISTS ("+
    "SELECT 1 FROM unnest($4::text[]) AS wanted(feature) "+
    "WHERE NOT EXISTS (SELECT 1 FROM public.resource_features f "+
    "WHERE f.resource_id=r.id AND LOWER(f.feature)=wanted.feature)) "+
  "ORDER BY r.name,r.id"
 rows,err:=s.db.Query(ctx,query,kind,from,to,features)
 if err!=nil {return nil,fmt.Errorf("search resources: %w",err)}
 defer rows.Close()
 result:=make([]models.Resource,0)
 for rows.Next(){
  item,err:=scanResource(rows)
  if err!=nil {return nil,fmt.Errorf("scan resource search: %w",err)}
  result=append(result,item)
 }
 if err:=rows.Err();err!=nil {return nil,fmt.Errorf("iterate resource search: %w",err)}
 return result,nil
}
