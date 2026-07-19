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
  mustChangePassword: boolean;
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

export interface DomainCheckResult {
  name: string;
  status: "pass" | "warn" | "fail";
  detail: string;
}

export interface DomainHealthResult {
  domain: string;
  status: "healthy" | "degraded" | "unhealthy";
  checks: DomainCheckResult[];
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

export interface MailThread {
  threadId: string;
  subject: string;
  messages: MailMessage[];
  lastDate: string;
  participantCount: number;
  unreadCount: number;
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

export interface ThreadsResponse {
  threads: MailThread[];
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
  status: string;
  email: string;
  role?: string;
  mustChangePassword: boolean;
}

export interface AuthMeResponse {
  email: string;
  role?: string;
  mustChangePassword: boolean;
}

export interface LogoutResponse {
  status: string;
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

export interface VacationSetRequest {
  subject: string;
  body: string;
  enabled: boolean;
}

export interface VacationResponse {
  email: string;
  subject?: string;
  body?: string;
  enabled: boolean;
  status?: string;
}

export interface SignatureResponse {
  email: string;
  content: string;
  enabled: boolean;
}

export interface UpsertSignatureRequest {
  content: string;
  enabled: boolean;
}

export interface DailyStat {
  date: string;
  sent: number;
  received: number;
  bounced: number;
  spamCaught: number;
}

export interface StatsSummary {
  totalSent: number;
  totalReceived: number;
  totalBounced: number;
  totalSpamCaught: number;
}

export interface MailQueueStatus {
  total: number;
  deferred: number;
  active: number;
  oldestAge: string;
}

export interface UserImportError {
  row: number;
  email?: string;
  error: string;
}

export interface UserImportResult {
  created: number;
  skipped: number;
  errors: UserImportError[];
}
