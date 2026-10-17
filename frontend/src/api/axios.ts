import axios from 'axios'
const api=axios.create({baseURL:'/api',timeout:15000})
api.interceptors.request.use(config=>{const token=localStorage.getItem('sikapj_token'); if(token) config.headers.Authorization=`Bearer ${token}`; return config})
api.interceptors.response.use(r=>r,error=>{
 const status=error.response?.status
 if(status===401 && error.config?.url !== '/auth/password' && error.config?.url !== '/auth/login'){localStorage.removeItem('sikapj_token');localStorage.removeItem('sikapj_user');sessionStorage.setItem('sikapj_notice','Sesi Anda berakhir. Silakan masuk kembali.');if(location.pathname!='/login') location.assign('/login')}
 if(status===403 && error.response?.data?.code==='PASSWORD_CHANGE_REQUIRED'){if(location.pathname!='/change-password') location.assign('/change-password')}
 else if(status===403){window.dispatchEvent(new CustomEvent('sikapj:forbidden',{detail:'Anda tidak memiliki izin untuk tindakan tersebut.'}))}
 return Promise.reject(error)
})
export default api
