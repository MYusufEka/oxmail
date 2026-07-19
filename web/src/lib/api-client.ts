import type {
  Contact,
  DailyStat,
  Domain,
  DomainHealthResult,
  MailQueueStatus,
  StatsSummary,
  User,
  Alias,
  DKIMKey,
  DNSRecord,
  DNSCheckResult,
  HealthStatus,
  LogEntry,
  LogsResponse,
  MailMessage,
  MailFolder,
  InboxResponse,
  FoldersResponse,
  ThreadsResponse,
  PaginatedResponse,
  CreateContactRequest,
  CreateDomainRequest,
  CreateUserRequest,
  UpdateContactRequest,
  UpdateUserRequest,
  CreateAliasRequest,
  UpdateAliasRequest,
  SendMailRequest,
  SendMailResponse,
  LoginRequest,
  LoginResponse,
  AuthMeResponse,
  LogoutResponse,
  ChangePasswordRequest,
  ChangePasswordResponse,
  ErrorResponse,
  SieveResponse,
  SignatureResponse,
  UpsertSignatureRequest,
  UserImportResult,
  VacationResponse,
  VacationSetRequest,
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
  const optionHeaders = new Headers(options?.headers);
  const isFormData = options?.body instanceof FormData;

  if (!isFormData && !optionHeaders.has("Content-Type")) {
    optionHeaders.set("Content-Type", "application/json");
  }

  const response = await fetch(url, {
    ...options,
    credentials: "include",
    headers: optionHeaders,
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

  getDomainHealth(name: string): Promise<DomainHealthResult> {
    return request(`/api/domains/${encodeURIComponent(name)}/health`);
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

  updateUser(domainId: number, userId: number, payload: UpdateUserRequest): Promise<User> {
    return request(`/api/domains/${domainId}/users/${userId}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    });
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

  updateAlias(domainId: number, aliasId: number, payload: UpdateAliasRequest): Promise<Alias> {
    return request(`/api/domains/${domainId}/aliases/${aliasId}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    });
  },

  // DKIM
  getDkim(domain: string): Promise<DKIMKey> {
    return request(`/api/domains/${domain}/dkim/`);
  },

  generateDkim(domain: string): Promise<DKIMKey> {
    return request(`/api/domains/${domain}/dkim/`, { method: "POST" });
  },

  // Health
  getHealth(): Promise<HealthStatus> {
    return request("/api/health");
  },

  // Logs
  getLogs(params?: PaginationParams): Promise<LogsResponse> {
    return request(`/api/logs${buildQuery(params)}`);
  },

  // Mail
  getInbox(userId: number, params?: PaginationParams, _email?: string): Promise<InboxResponse> {
    return request(`/api/mail/${userId}/inbox${buildQuery(params)}`);
  },

  getMessage(userId: number, messageId: number, _email?: string): Promise<MailMessage> {
    return request(`/api/mail/${userId}/messages/${messageId}`);
  },

  sendMail(payload: SendMailRequest): Promise<SendMailResponse> {
    return request(`/api/mail/send?auth_user=${encodeURIComponent(payload.from)}`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  markAsRead(userId: number, messageId: number, _email?: string): Promise<void> {
    return request(`/api/mail/${userId}/messages/${messageId}`, {
      method: "PATCH",
      body: JSON.stringify({ read: true }),
    });
  },

  toggleRead(userId: number, messageId: number, _email?: string): Promise<void> {
    return request(`/api/mail/${userId}/messages/${messageId}/toggle-read`, {
      method: "PATCH",
    });
  },

  trashMessage(userId: number, messageId: number, _email?: string): Promise<void> {
    return request(`/api/mail/${userId}/messages/${messageId}`, {
      method: "DELETE",
    });
  },

  searchMail(query: string, user: string): Promise<MailMessage[]> {
    const params = new URLSearchParams({ q: query, user });
    return request(`/api/mail/search?${params.toString()}`);
  },

  getMailFolders(user: string): Promise<FoldersResponse> {
    return request(`/api/mail/folders?user=${encodeURIComponent(user)}`);
  },

  getFolderMessages(folder: string, user: string, params?: PaginationParams): Promise<InboxResponse> {
    const base = buildQuery(params);
    const userQuery = `user=${encodeURIComponent(user)}`;
    const query = base ? `${base}&${userQuery}` : `?${userQuery}`;
    return request(`/api/mail/folders/${encodeURIComponent(folder)}/messages${query}`);
  },

  getThreads(folder: string, user: string, params?: PaginationParams): Promise<ThreadsResponse> {
    const base = buildQuery(params);
    const userQuery = `user=${encodeURIComponent(user)}`;
    const query = base ? `${base}&${userQuery}` : `?${userQuery}`;
    return request(`/api/mail/folders/${encodeURIComponent(folder)}/threads${query}`);
  },

  // Auth
  login(payload: LoginRequest): Promise<LoginResponse> {
    return request("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  me(): Promise<AuthMeResponse> {
    return request("/api/auth/me");
  },

  logout(): Promise<LogoutResponse> {
    return request("/api/auth/logout", { method: "POST" });
  },

  changePassword(payload: ChangePasswordRequest): Promise<ChangePasswordResponse> {
    return request("/api/auth/change-password", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  // DNS
  getDnsRecords(): Promise<{ records: DNSRecord[] }> {
    return request("/api/dns/records");
  },

  getDnsCheck(): Promise<{ results: DNSCheckResult[] }> {
    return request("/api/dns/check");
  },

  // Sieve filters
  getSieveScript(email: string): Promise<SieveResponse> {
    return request(`/api/mail/filters/${encodeURIComponent(email)}`);
  },

  setSieveScript(email: string, script: string): Promise<SieveResponse> {
    return request(`/api/mail/filters/${encodeURIComponent(email)}`, {
      method: "POST",
      body: JSON.stringify({ script }),
    });
  },

  deleteSieveScript(email: string): Promise<SieveResponse> {
    return request(`/api/mail/filters/${encodeURIComponent(email)}`, {
      method: "DELETE",
    });
  },

  // Vacation scripts
  getVacationScript(email: string): Promise<SieveResponse> {
    return request(`/api/mail/vacation/${encodeURIComponent(email)}`);
  },

  setVacation(email: string, payload: VacationSetRequest): Promise<VacationResponse> {
    return request(`/api/mail/vacation/${encodeURIComponent(email)}`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  setVacationScript(email: string, script: string): Promise<VacationResponse> {
    return apiClient.setVacation(email, {
      subject: "Vacation auto-reply",
      body: script,
      enabled: script !== "",
    });
  },

  deleteVacationScript(email: string): Promise<VacationResponse> {
    return request(`/api/mail/vacation/${encodeURIComponent(email)}`, {
      method: "DELETE",
    });
  },

  getSignature(email: string): Promise<SignatureResponse> {
    return request(`/api/mail/signature/${encodeURIComponent(email)}`);
  },

  upsertSignature(email: string, payload: UpsertSignatureRequest): Promise<SignatureResponse> {
    return request(`/api/mail/signature/${encodeURIComponent(email)}`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  deleteSignature(email: string): Promise<void> {
    return request(`/api/mail/signature/${encodeURIComponent(email)}`, {
      method: "DELETE",
    });
  },

  // Mail Folders
  createFolder(userEmail: string, folderName: string): Promise<{ status: string; name: string }> {
    return request("/api/mail/folders", {
      method: "POST",
      headers: { "X-Mail-User": userEmail },
      body: JSON.stringify({ name: folderName }),
    });
  },

  deleteFolder(userEmail: string, folderName: string): Promise<{ status: string }> {
    return request(`/api/mail/folders/${encodeURIComponent(folderName)}`, {
      method: "DELETE",
      headers: { "X-Mail-User": userEmail },
    });
  },

  renameFolder(userEmail: string, oldName: string, newName: string): Promise<{ status: string; name: string }> {
    return request(`/api/mail/folders/${encodeURIComponent(oldName)}`, {
      method: "PATCH",
      headers: { "X-Mail-User": userEmail },
      body: JSON.stringify({ new_name: newName }),
    });
  },

  moveMessage(userEmail: string, uid: number, fromFolder: string, toFolder: string): Promise<{ status: string }> {
    return request(`/api/mail/messages/${uid}/move`, {
      method: "POST",
      headers: { "X-Mail-User": userEmail },
      body: JSON.stringify({ from_folder: fromFolder, to_folder: toFolder }),
    });
  },

  // Stats
  getStats(days: number): Promise<DailyStat[]> {
    return request(`/api/stats?days=${days}`);
  },

  getStatsSummary(): Promise<StatsSummary> {
    return request("/api/stats/summary");
  },

  // Mail queue
  getMailQueue(): Promise<MailQueueStatus> {
    return request("/api/mail/queue");
  },

  importUsers(domainId: number, file: File): Promise<UserImportResult> {
    const formData = new FormData();
    formData.append("file", file);
    return request(`/api/domains/${domainId}/users/import`, {
      method: "POST",
      body: formData,
      headers: {},
    });
  },

  // Contacts
  getContacts(userEmail: string): Promise<Contact[]> {
    return request(`/api/contacts/${encodeURIComponent(userEmail)}`);
  },

  createContact(userEmail: string, payload: CreateContactRequest): Promise<Contact> {
    return request(`/api/contacts/${encodeURIComponent(userEmail)}`, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  updateContact(userEmail: string, contactId: number, payload: UpdateContactRequest): Promise<Contact> {
    return request(`/api/contacts/${encodeURIComponent(userEmail)}/${contactId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  },

  deleteContact(userEmail: string, contactId: number): Promise<void> {
    return request(`/api/contacts/${encodeURIComponent(userEmail)}/${contactId}`, {
      method: "DELETE",
    });
  },
} as const;
