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
  services: {
    postfix: ServiceHealth;
    dovecot: ServiceHealth;
    rspamd: ServiceHealth;
    redis: ServiceHealth;
  };
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

export interface CreateAliasRequest {
  sourceAddress: string;
  destinationAddress: string;
}

export interface SendMailRequest {
  from: string;
  to: string[];
  cc?: string[];
  subject: string;
  bodyText?: string;
  bodyHtml?: string;
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
