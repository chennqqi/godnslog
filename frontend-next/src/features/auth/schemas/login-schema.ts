import { z } from 'zod'

/** Zod schema for login form */
export const loginSchema = z.object({
  username: z.string().min(1, 'Username is required').max(100, 'Username must be less than 100 characters'),
  password: z.string().min(1, 'Password is required').max(200, 'Password must be less than 200 characters'),
})

export type LoginFormValues = z.infer<typeof loginSchema>
