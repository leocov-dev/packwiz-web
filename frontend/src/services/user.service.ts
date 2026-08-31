import {User, UserListResponse} from "@/interfaces/user";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";
import type {UserProfileFormData} from "@/components/user/UserProfile.vue";
import type {CreateUserFormData} from "@/components/user/AdminUserCreateForm.vue";


export async function userLogin(username: string, password: string): Promise<User> {
  const response = await apiClient.postForm('v1/auth/login', {username, password})
  return plainToInstance(User, response.data)
}


export async function userLogout(): Promise<void> {
  return apiClient.post('v1/auth/logout')
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

export async function fetchUserById(id: number): Promise<User> {
  const response = await apiClient.get(`v1/admin/users/${id}`)
  return plainToInstance(User, response.data)
}

export async function updateUserById(id: number, userData: UserProfileFormData): Promise<void> {
  await apiClient.patch(`v1/admin/users/${id}`, userData)
}

export async function createUser(userData: CreateUserFormData): Promise<User> {
  const response = await apiClient.post('v1/admin/users', userData)
  return plainToInstance(User, response.data)
}

export async function deactivateUserById(id: number): Promise<void> {
  await apiClient.patch(`v1/admin/users/${id}/deactivate`)
}

export async function reactivateUserById(id: number): Promise<void> {
  await apiClient.patch(`v1/admin/users/${id}/reactivate`)
}

export async function fetchUsersPaginated(
  page: number,
  pageSize: number,
  nameSearch = '',
  emailSearch = '',
  userType = '',
): Promise<UserListResponse> {
  const params = new URLSearchParams({page: String(page), pageSize: String(pageSize)})
  if (nameSearch) params.append('nameSearch', nameSearch)
  if (emailSearch) params.append('emailSearch', emailSearch)
  if (userType) params.append('userType', userType)

  const response = await apiClient.get(`v1/admin/users?${params}`)
  return plainToInstance(UserListResponse, response.data)
}
