const empty = () => Object.fromEntries(['employees','clients','vehicles','users','receptions','diagnoses','quotes','orders','payments','contacts','rescues','suppliers','roles','services','spareParts','units'].map(k=>[k,[]]));
let state=empty();
export const realStore=()=>structuredClone(state);
export function replaceRealStore(next) { state={...empty(),...next}; }
export function updateRealStore(action) { action(state); }
export const clearRealStore=()=>{state=empty();};
