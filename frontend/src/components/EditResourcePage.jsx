import {useEffect,useState} from 'react'
import {Link,useParams} from 'react-router-dom'
import {Button,Card,Field,SelectField} from './ui.jsx'
import {useAuth} from '../AuthContext.jsx'
import {paths} from '../routes.js'
import {validateResource} from '../resource-admin-api.js'
import {loadAdminResource,saveResourceEdit} from '../resource-edit-api.js'
const types=[{value:'DESK',label:'Delovna miza'},{value:'MEETING_ROOM',label:'Sejna soba'},{value:'PARKING_SPOT',label:'Parkirišče'}]
export default function EditResourcePage(){
 const {id}=useParams()
 const {token,logout}=useAuth()
 const [values,setValues]=useState(null)
 const [errors,setErrors]=useState({})
 const [feedback,setFeedback]=useState('')
 const [busy,setBusy]=useState(false)
 useEffect(()=>{
  const controller=new AbortController();let active=true
  loadAdminResource(id,{token,logout,apiBaseUrl:import.meta.env.VITE_API_BASE_URL||'http://localhost:8080/api',
   fetchImpl:(url,opts)=>fetch(url,{...opts,signal:controller.signal})})
   .then(data=>{if(active)setValues({name:data.resource.name,type:data.resource.type,
     floor:data.resource.floor==null?'':String(data.resource.floor),
     capacity:String(data.resource.capacity),is_active:data.resource.is_active,
     features:data.features.join(', ')})})
   .catch(e=>{if(active&&e.name!=='AbortError')setFeedback(e.message)})
  return ()=>{active=false;controller.abort()}
 },[id,token,logout])
 const update=(key,value)=>{setValues(current=>({...current,[key]:value}));setErrors(current=>({...current,[key]:undefined}));setFeedback('')}
 async function submit(event){
  event.preventDefault();if(busy||!values)return
  const problems=validateResource(values);setErrors(problems)
  if(Object.keys(problems).length)return
  setBusy(true)
  try{await saveResourceEdit(id,values,{token,logout,apiBaseUrl:import.meta.env.VITE_API_BASE_URL||'http://localhost:8080/api'})
   setFeedback('Spremembe in oprema so shranjene.')}
  catch(e){setFeedback(e.message||'Shranjevanje ni uspelo.')}
  finally{setBusy(false)}
 }
 return <Card className="auth-card">
  <span className="eyebrow">Upravljanje virov</span>
  <h1>Uredi vir</h1>
  <p><Link to={paths.adminResources}>Nazaj na seznam</Link></p>
  {feedback&&<p role="status" className="auth-feedback">{feedback}</p>}
  {!values&&<p role="status">Nalagam podatke vira ...</p>}
  {values&&<form className="auth-form" onSubmit={submit} noValidate>
   <Field label="Naziv" value={values.name} required maxLength={150} onChange={e=>update('name',e.target.value)} aria-invalid={!!errors.name}/>
   {errors.name&&<span role="alert" className="form-error">{errors.name}</span>}
   <SelectField label="Vrsta" options={types} value={values.type} onChange={e=>update('type',e.target.value)}/>
   <Field label="Nadstropje" type="number" min={0} step={1} value={values.floor} onChange={e=>update('floor',e.target.value)}/>
   {errors.floor&&<span role="alert" className="form-error">{errors.floor}</span>}
   <Field label="Kapaciteta" type="number" required min={1} step={1} value={values.capacity} onChange={e=>update('capacity',e.target.value)}/>
   {errors.capacity&&<span role="alert" className="form-error">{errors.capacity}</span>}
   <Field label="Oprema (ločena z vejicami)" value={values.features} onChange={e=>update('features',e.target.value)}/>
   {errors.features&&<span role="alert" className="form-error">{errors.features}</span>}
   <label>Aktiven <input type="checkbox" checked={values.is_active} onChange={e=>update('is_active',e.target.checked)}/></label>
   <Button type="submit" disabled={busy}>{busy?'Shranjujem ...':'Shrani spremembe'}</Button>
  </form>}
 </Card>
}
