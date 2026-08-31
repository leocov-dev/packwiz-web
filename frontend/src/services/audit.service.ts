import {AuditListResponse} from "@/interfaces/audit";
import {apiClient} from "@/services/api.service";
import {plainToInstance} from "class-transformer";

export async function fetchAuditsPaginated(
  page: number,
  pageSize: number,
  action = '',
  userId?: number,
  startDate = '',
  endDate = '',
): Promise<AuditListResponse> {
  const params = new URLSearchParams({page: String(page), pageSize: String(pageSize)})
  if (action) params.append('action', action)
  if (userId) params.append('userId', String(userId))
  if (startDate) params.append('startDate', startDate)
  if (endDate) params.append('endDate', endDate)

  const response = await apiClient.get(`v1/admin/audits?${params}`)
  return plainToInstance(AuditListResponse, response.data)
}
