export interface Domain {
  id: number;
  name: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: number;
  email: string;
  domainId: number;
  displayName?: string;
  quota: number;
  storageUsed?: number;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Alias {
  id: number;
  sourceAddress: string;
  destinationAddress: string;
  active: boolean;
  createdAt: string;
}

export interface DKIMKey {
  domain: string;
  selector: string;
  publicKey: string;
  dnsRecord?: string;
  createdAt: string;
}

export interface HealthStatus {
  status: "healthy" | "degraded" | "unhealthy";
  version: string;
  uptime: string;
  services: ServiceHealthEntry[];
}

export interface ServiceHealthEntry {
  name: string;
  status: "healthy" | "unhealthy";
  latencyMs: number;
}

export interface ServiceHealth {
  status: "up" | "down";
  latencyMs?: number;
}

export interface LogEntry {
  id: number;
  timestamp: string;
  service: "postfix" | "dovecot" | "rspamd" | "api";
  level: "debug" | "info" | "warn" | "error";
  message: string;
}

export interface LogsResponse {
  entries: LogEntry[];
  total: number;
  limit: number;
  offset: number;
}

export interface MailMessage {
  id: number;
  from: string;
  to: string[];
  cc?: string[];
  subject: string;
  bodyText?: string;
  bodyHtml?: string;
  read: boolean;
  receivedAt: string;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: Pagination;
}

export interface InboxResponse {
  messages: MailMessage[];
  pagination: Pagination;
}

export interface MailFolder {
  name: string;
  delimiter: string;
  unread: number;
  total: number;
}

export interface FoldersResponse {
  folders: MailFolder[];
}

export interface DNSRecord {
  domain: string;
  type: "MX" | "TXT" | "CNAME" | "A";
  name: string;
  value: string;
  priority?: number;
}

export interface DNSCheckResult {
  domain: string;
  record: string;
  expected: string;
  actual: string;
  valid: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ChangePasswordRequest {
  email: string;
  currentPassword: string;
  newPassword: string;
}

export interface ChangePasswordResponse {
  status: "password_changed";
}

export interface LoginResponse {
  token: string;
  expiresAt: string;
}

export interface CreateDomainRequest {
  name: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  displayName?: string;
  quota?: number;
}

export interface UpdateUserRequest {
  password?: string;
  displayName?: string;
  quota?: number;
}

export interface CreateAliasRequest {
  sourceAddress: string;
  destinationAddress: string;
}

export interface UpdateAliasRequest {
  sourceAddress: string;
  destinationAddress: string;
}

export interface MailAttachment {
  filename: string;
  content: string;
  mimeType: string;
}

export interface SendMailRequest {
  from: string;
  to: string[];
  cc?: string[];
  subject: string;
  bodyText?: string;
  bodyHtml?: string;
  attachments?: MailAttachment[];
}

export interface SendMailResponse {
  messageId: string;
  status: "queued";
}

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}

export interface Contact {
  id: number;
  userEmail: string;
  name: string;
  email: string;
  phone?: string;
  createdAt: string;
}

export interface CreateContactRequest {
  name: string;
  email: string;
  phone?: string;
}

export interface UpdateContactRequest {
  name?: string;
  email?: string;
  phone?: string;
}

export interface SieveResponse {
  email: string;
  script?: string;
  active: boolean;
  status?: string;
}

export interface SieveSetRequest {
  script: string;
}
