import { z } from "zod";

export const paginationSchema = z.object({
  page: z.number().int().min(1),
  limit: z.number().int().min(1).max(100),
  total: z.number().int().min(0),
});

export const errorResponseSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
  }),
});

// Domain schemas
export const domainSchema = z.object({
  id: z.number().int(),
  name: z.string().min(3),
  active: z.boolean(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const createDomainRequestSchema = z.object({
  name: z.string().min(3),
});

export const domainListResponseSchema = z.object({
  data: z.array(domainSchema),
  pagination: paginationSchema,
});

// User schemas
export const userSchema = z.object({
  id: z.number().int(),
  email: z.string().email(),
  domainId: z.number().int(),
  displayName: z.string().optional(),
  quota: z.number().int().min(0),
  active: z.boolean(),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const createUserRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
  displayName: z.string().optional(),
  quota: z.number().int().min(0).optional(),
});

export const userListResponseSchema = z.object({
  data: z.array(userSchema),
  pagination: paginationSchema,
});

// Alias schemas
export const aliasSchema = z.object({
  id: z.number().int(),
  sourceAddress: z.string().email(),
  destinationAddress: z.string().email(),
  active: z.boolean(),
  createdAt: z.string().datetime(),
});

export const createAliasRequestSchema = z.object({
  sourceAddress: z.string().email(),
  destinationAddress: z.string().email(),
});

export const aliasListResponseSchema = z.object({
  data: z.array(aliasSchema),
  pagination: paginationSchema,
});

// DKIM schemas
export const dkimKeySchema = z.object({
  domain: z.string(),
  selector: z.string(),
  publicKey: z.string(),
  dnsRecord: z.string().optional(),
  createdAt: z.string().datetime(),
});

// Health schemas
export const serviceHealthSchema = z.object({
  status: z.enum(["up", "down"]),
  latencyMs: z.number().int().optional(),
});

export const healthStatusSchema = z.object({
  status: z.enum(["healthy", "degraded", "unhealthy"]),
  services: z.object({
    postfix: serviceHealthSchema,
    dovecot: serviceHealthSchema,
    rspamd: serviceHealthSchema,
    redis: serviceHealthSchema,
  }),
});

// Log schemas
export const logEntrySchema = z.object({
  id: z.number().int(),
  timestamp: z.string().datetime(),
  service: z.enum(["postfix", "dovecot", "rspamd", "api"]),
  level: z.enum(["debug", "info", "warn", "error"]),
  message: z.string(),
});

// Mail schemas
export const mailMessageSchema = z.object({
  id: z.number().int(),
  from: z.string().email(),
  to: z.array(z.string().email()).min(1),
  cc: z.array(z.string().email()).optional(),
  subject: z.string(),
  bodyText: z.string().optional(),
  bodyHtml: z.string().optional(),
  read: z.boolean(),
  receivedAt: z.string().datetime(),
});

export const mailMessageListResponseSchema = z.object({
  data: z.array(mailMessageSchema),
  pagination: paginationSchema,
});

export const mailAttachmentSchema = z.object({
  filename: z.string(),
  content: z.string(),
  mimeType: z.string(),
});

export const sendMailRequestSchema = z.object({
  from: z.string().email(),
  to: z.array(z.string().email()).min(1),
  cc: z.array(z.string().email()).optional(),
  subject: z.string().min(1),
  bodyText: z.string().optional(),
  bodyHtml: z.string().optional(),
  attachments: z.array(mailAttachmentSchema).optional(),
});

export const sendMailResponseSchema = z.object({
  messageId: z.string(),
  status: z.enum(["queued"]),
});

// DNS schemas
export const dnsRecordSchema = z.object({
  domain: z.string(),
  type: z.enum(["MX", "TXT", "CNAME", "A"]),
  name: z.string(),
  value: z.string(),
  priority: z.number().int().optional(),
});

export const dnsCheckResultSchema = z.object({
  domain: z.string(),
  record: z.string(),
  expected: z.string(),
  actual: z.string(),
  valid: z.boolean(),
});

// Auth schemas
export const loginRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
});

export const loginResponseSchema = z.object({
  token: z.string(),
  expiresAt: z.string().datetime(),
});

export const changePasswordRequestSchema = z.object({
  email: z.string().email("Enter a valid email address"),
  currentPassword: z.string().min(1, "Current password is required"),
  newPassword: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .regex(/[A-Z]/, "Must contain at least one uppercase letter")
    .regex(/[0-9]/, "Must contain at least one number"),
});

export const changePasswordResponseSchema = z.object({
  status: z.literal("password_changed"),
});
