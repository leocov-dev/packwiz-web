import {User} from "@/interfaces/user";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";


export async function userLogin(username: string, password: string): Promise<User> {
  return apiClient.postForm('v1/login', {username, password})
}


export async function userLogout(): Promise<void> {
  return apiClient.post('v1/logout')
}


export async function getCurrentUser(): Promise<User> {
  const response = await apiClient.get('v1/user/current')
  return plainToInstance(User, response.data)
}
