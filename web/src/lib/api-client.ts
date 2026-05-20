import type {
  Domain,
  User,
  Alias,
  DKIMKey,
  DNSRecord,
  DNSCheckResult,
  HealthStatus,
  LogEntry,
  MailMessage,
  PaginatedResponse,
  CreateDomainRequest,
  CreateUserRequest,
  CreateAliasRequest,
  SendMailRequest,
  SendMailResponse,
  ErrorResponse,
} from "@/types/api";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${path}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    let errorBody: ErrorResponse | null = null;
    try {
      errorBody = (await response.json()) as ErrorResponse;
    } catch {
      // Response body is not JSON
    }

    throw new ApiError(
      response.status,
      errorBody?.error.code ?? "UNKNOWN_ERROR",
      errorBody?.error.message ?? `Request failed with status ${response.status}`,
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

function buildQuery(params?: PaginationParams): string {
  if (!params) return "";
  const searchParams = new URLSearchParams();
  if (params.page !== undefined) searchParams.set("page", String(params.page));
  if (params.limit !== undefined) searchParams.set("limit", String(params.limit));
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export const apiClient = {
  // Domains
  getDomains(params?: PaginationParams): Promise<PaginatedResponse<Domain>> {
    return request(`/api/domains${buildQuery(params)}`);
  },

  createDomain(payload: CreateDomainRequest): Promise<Domain> {
    return request("/api/domains", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  deleteDomain(domainId: number): Promise<void> {
    return request(`/api/domains/${domainId}`, { method: "DELETE" });
  },

  // Users
  getUsers(domainId: number, params?: PaginationParams): Promise<PaginatedResponse<User>> {
    return request(`/api/domains/${domainId}/users${buildQuery(params)}`);
  },

  createUser(domainId: number, payload: CreateUserRequest): Promise<User> {
    return request(`/api/domains/${domainId}/users`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  deleteUser(domainId: number, userId: number): Promise<void> {
    return request(`/api/domains/${domainId}/users/${userId}`, { method: "DELETE" });
  },

  // Aliases
  getAliases(domainId: number, params?: PaginationParams): Promise<PaginatedResponse<Alias>> {
    return request(`/api/domains/${domainId}/aliases${buildQuery(params)}`);
  },

  createAlias(domainId: number, payload: CreateAliasRequest): Promise<Alias> {
    return request(`/api/domains/${domainId}/aliases`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  deleteAlias(domainId: number, aliasId: number): Promise<void> {
    return request(`/api/domains/${domainId}/aliases/${aliasId}`, { method: "DELETE" });
  },

  // DKIM
  getDkim(domain: string): Promise<DKIMKey> {
    return request(`/api/dkim/${domain}`);
  },

  generateDkim(domain: string): Promise<DKIMKey> {
    return request(`/api/dkim/${domain}`, { method: "POST" });
  },

  // Health
  getHealth(): Promise<HealthStatus> {
    return request("/api/health");
  },

  // Logs
  getLogs(params?: PaginationParams): Promise<PaginatedResponse<LogEntry>> {
    return request(`/api/logs${buildQuery(params)}`);
  },

  // Mail
  getInbox(userId: number, params?: PaginationParams): Promise<PaginatedResponse<MailMessage>> {
    return request(`/api/mail/${userId}/inbox${buildQuery(params)}`);
  },

  getMessage(userId: number, messageId: number): Promise<MailMessage> {
    return request(`/api/mail/${userId}/messages/${messageId}`);
  },

  sendMail(payload: SendMailRequest): Promise<SendMailResponse> {
    return request("/api/mail/send", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  markAsRead(userId: number, messageId: number): Promise<void> {
    return request(`/api/mail/${userId}/messages/${messageId}`, {
      method: "PATCH",
      body: JSON.stringify({ read: true }),
    });
  },

  toggleRead(userId: number, messageId: number): Promise<void> {
    return request(`/api/mail/${userId}/messages/${messageId}/toggle-read`, {
      method: "PATCH",
    });
  },

  trashMessage(userId: number, messageId: number): Promise<void> {
    return request(`/api/mail/${userId}/messages/${messageId}`, {
      method: "DELETE",
    });
  },

  searchMail(query: string, user: string): Promise<MailMessage[]> {
    const params = new URLSearchParams({ q: query, user });
    return request(`/api/mail/search?${params.toString()}`);
  },

  // DNS
  getDnsRecords(): Promise<{ records: DNSRecord[] }> {
    return request("/api/dns/records");
  },

  getDnsCheck(): Promise<{ results: DNSCheckResult[] }> {
    return request("/api/dns/check");
  },
} as const;
