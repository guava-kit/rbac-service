import { axios } from './api'

export interface SignupWithEmailData {
  username: string
  email: string
  password: string
}

export interface LoginWithUsername {
  username: string
  password: string
}

export async function Signup(
  type: 'email' | 'phone',
  data: SignupWithEmailData
) {
  return await axios.post(`/auth/signup?type=${type}`, data)
}

export async function Login(
  type: 'username' | 'email',
  data: LoginWithUsername
) {
  return await axios.post(`/auth/signin?type=${type}`, data)
}
