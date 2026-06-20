'use client'

import { useState, useCallback, type ReactNode } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface ConfirmOptions {
  title: string
  description?: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'default' | 'destructive'
}

interface ConfirmDialogState extends ConfirmOptions {
  open: boolean
  resolve: ((value: boolean) => void) | null
}

/** Hook that provides a confirm() function returning a Promise<boolean>, replacing native window.confirm */
export function useConfirmDialog() {
  const [state, setState] = useState<ConfirmDialogState>({
    open: false,
    title: '',
    resolve: null,
  })

  const confirm = useCallback((options: ConfirmOptions): Promise<boolean> => {
    return new Promise((resolve) => {
      setState({
        ...options,
        open: true,
        resolve,
      })
    })
  }, [])

  const handleConfirm = () => {
    state.resolve?.(true)
    setState((s) => ({ ...s, open: false, resolve: null }))
  }

  const handleCancel = () => {
    state.resolve?.(false)
    setState((s) => ({ ...s, open: false, resolve: null }))
  }

  const dialogElement: ReactNode = (
    <Dialog open={state.open} onOpenChange={(open) => { if (!open) handleCancel() }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{state.title}</DialogTitle>
          {state.description && (
            <DialogDescription>{state.description}</DialogDescription>
          )}
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={handleCancel}>
            {state.cancelLabel || 'Cancel'}
          </Button>
          <Button
            variant={state.variant === 'destructive' ? 'destructive' : 'default'}
            onClick={handleConfirm}
          >
            {state.confirmLabel || 'Confirm'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )

  return { confirm, dialogElement }
}
