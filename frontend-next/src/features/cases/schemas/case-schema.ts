import { z } from 'zod'

/** Zod schema for case creation/update */
export const caseSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200, 'Title must be less than 200 characters'),
  description: z.string().max(2000, 'Description must be less than 2000 characters'),
  target: z.string().max(255, 'Target must be less than 255 characters'),
  status: z.enum(['active', 'completed', 'archived']),
  tags: z.array(z.string()),
})

export type CaseFormValues = z.infer<typeof caseSchema>
