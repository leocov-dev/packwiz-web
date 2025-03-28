import {User} from "@/interfaces/user";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";
import type {UserProfileFormData} from "@/components/user/UserProfile.vue";


export async function userLogin(username: string, password: string): Promise<User> {
  return apiClient.postForm('v1/login', {username, password})
}


export async function userLogout(): Promise<void> {
  return apiClient.post('v1/logout')
}


export async function getCurrentUser(): Promise<User> {
  const response = await apiClient.get('v1/user')
  return plainToInstance(User, response.data)
}


export async function changePassword(oldPass: string, newPass: string): Promise<void> {
  await apiClient.postForm(
    'v1/user/password',
    {
      oldPassword: oldPass,
      newPassword: newPass,
    },
  )
}

export async function updateCurrentUser(userData: UserProfileFormData): Promise<void> {
  await apiClient.post(`v1/user/update`, userData)
}

export async function invalidateCurrentUserSessions(): Promise<void> {
  await apiClient.post('v1/user/invalidate-sessions')
}
