import {realMode,request,realSession,setSession,clearSession} from './api-service.js';
import {clearRealStore} from './real-state.js';
let refreshVersion=0;
export async function refreshSession(){if(!realMode)return getSession();const version=++refreshVersion;try{const response=await request('/auth/me');return version===refreshVersion?setSession(response):getSession();}catch(error){if(version!==refreshVersion)return getSession();if(error.status===401){clearRealStore();return null;}throw error;}}
import {getStore} from './store-service.js';
import {passwordDigest} from './workshop-service.js';
const SESSION_KEY='bravo-jac:demo-session:v1',REMEMBER_KEY='bravo-jac:remembered-user:v1';
export const demoCredentials=Object.freeze({username:'admin',password:'Bravo2026!'});
export function getSession(){if(realMode)return realSession();try{const value=JSON.parse(sessionStorage.getItem(SESSION_KEY));if(!value?.demo)return null;const u=getStore().users.find(u=>(u.id===value.id||u.username===value.username)&&u.active);return u?{id:u.id,name:u.name,username:u.username,role:u.role,demo:true}:null;}catch{return null;}}
export function getRememberedUser(){try{return localStorage.getItem(realMode?'bravo-jac:real-remembered':REMEMBER_KEY)||'';}catch{return '';}}
export async function login({username,password,remember}){
 ++refreshVersion;
 if(realMode){const session=setSession(await request('/auth/login',{method:'POST',body:{username:username.trim(),password}}));clearRealStore();try{if(remember)localStorage.setItem('bravo-jac:real-remembered',session.username);else localStorage.removeItem('bravo-jac:real-remembered');}catch{}return session;}
 const user=getStore().users.find(u=>(u.username===username.trim().toLowerCase()||u.email===username.trim().toLowerCase())&&u.active);
 const defaultPassword=user?.id==='u-kevin'?'Auxiliar2026!':'Bravo2026!';
 const valid=user&&(user.passwordDigest?await passwordDigest(password)===user.passwordDigest:password===defaultPassword);
 if(!valid)throw new Error('El usuario o la contraseña no son correctos, o el usuario está inactivo.');
 const session={id:user.id,name:user.name,username:user.username,role:user.role,demo:true};
 try{sessionStorage.setItem(SESSION_KEY,JSON.stringify(session));if(remember)localStorage.setItem(REMEMBER_KEY,user.username);else localStorage.removeItem(REMEMBER_KEY);}catch{throw new Error('Habilita el almacenamiento del navegador para continuar.');}return session;
}
export function logout(){++refreshVersion;if(realMode)return request('/auth/logout',{method:'POST',body:{}}).then(()=>{clearSession();clearRealStore();});sessionStorage.removeItem(SESSION_KEY);}
