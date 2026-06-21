'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '@/features/users/hooks/use-users'
import { useConfirmDialog } from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

interface User {
  id: string
  username: string
  email: string
  role: number
  created_at: string
}

/** Maps backend role ints (models/api.go) to a short label */
function roleLabel(role: number): string {
  switch (role) {
    case 0:
      return 'Super admin'
    case 1:
      return 'Admin'
    case 2:
      return 'User'
    case 3:
      return 'Guest'
    default:
      return `Role ${role}`
  }
}

export default function UsersPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const { data: usersData, isLoading: loading } = useUsers()
  const createUser = useCreateUser()
  const updateUser = useUpdateUser()
  const deleteUser = useDeleteUser()
  const users = usersData?.data?.items ?? []
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [newUser, setNewUser] = useState({ username: '', email: '', password: '', role: 2 })
  const [editForm, setEditForm] = useState({ email: '', role: 2, password: '' })
  const [submitting, setSubmitting] = useState(false)
  const { confirm, dialogElement } = useConfirmDialog()

  const handleCreateUser = async () => {
    if (!newUser.username || !newUser.password) return
    setSubmitting(true)
    try {
      await createUser.mutateAsync(newUser)
      setShowCreateModal(false)
      setNewUser({ username: '', email: '', password: '', role: 2 })
    } catch (error) {
      console.error('Failed to create user:', error)
    } finally {
      setSubmitting(false)
    }
  }

  const handleEditUser = (user: User) => {
    setEditingUser(user)
    setEditForm({ email: user.email, role: user.role, password: '' })
    setShowEditModal(true)
  }

  const handleUpdateUser = async () => {
    if (!editingUser) return
    setSubmitting(true)
    try {
      const data: { email?: string; role?: number; password?: string } = {
        email: editForm.email,
        role: editForm.role,
      }
      if (editForm.password) {
        data.password = editForm.password
      }
      await updateUser.mutateAsync({ id: editingUser.id, data })
      setShowEditModal(false)
      setEditingUser(null)
    } catch (error) {
      console.error('Failed to update user:', error)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDeleteUser = async (user: User) => {
    const ok = await confirm({
      title: 'Delete User',
      description: `Are you sure you want to delete user "${user.username}"? This action cannot be undone.`,
      confirmLabel: 'Delete',
      variant: 'destructive',
    })
    if (!ok) return
    try {
      await deleteUser.mutateAsync(user.id)
    } catch (error) {
      console.error('Failed to delete user:', error)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading...</p>
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">User Management</h2>

      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">User List</h3>
            <Button onClick={() => setShowCreateModal(true)}>
              Create User
            </Button>
          </div>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Username</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-gray-500">
                    No users
                  </TableCell>
                </TableRow>
              ) : (
                users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>{user.username}</TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>
                      <Badge variant={user.role === 0 || user.role === 1 ? 'default' : 'outline'}>
                        {roleLabel(user.role)}
                      </Badge>
                    </TableCell>
                    <TableCell>{new Date(user.created_at).toLocaleString()}</TableCell>
                    <TableCell>
                      <Button variant="ghost" size="sm" className="mr-2" onClick={() => handleEditUser(user)}>
                        Edit
                      </Button>
                      <Button variant="destructive" size="sm" onClick={() => handleDeleteUser(user)}>
                        Delete
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      {/* Create User Modal */}
      <Dialog open={showCreateModal} onOpenChange={setShowCreateModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create User</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                value={newUser.username}
                onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={newUser.email}
                onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={newUser.password}
                onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="role">Role</Label>
              <Select value={String(newUser.role)} onValueChange={(value) => setNewUser({ ...newUser, role: parseInt(value) })}>
                <SelectTrigger id="role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="2">User</SelectItem>
                  <SelectItem value="1">Admin</SelectItem>
                  <SelectItem value="3">Guest</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={handleCreateUser} disabled={submitting}>
              {submitting ? 'Creating...' : 'Create'}
            </Button>
            <Button variant="outline" onClick={() => setShowCreateModal(false)}>
              Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit User Modal */}
      <Dialog open={showEditModal} onOpenChange={setShowEditModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User{editingUser ? `: ${editingUser.username}` : ''}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="edit-email">Email</Label>
              <Input
                id="edit-email"
                type="email"
                value={editForm.email}
                onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
              />
            </div>
            <div>
              <Label htmlFor="edit-role">Role</Label>
              <Select value={String(editForm.role)} onValueChange={(value) => setEditForm({ ...editForm, role: parseInt(value) })}>
                <SelectTrigger id="edit-role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">Super admin</SelectItem>
                  <SelectItem value="1">Admin</SelectItem>
                  <SelectItem value="2">User</SelectItem>
                  <SelectItem value="3">Guest</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="edit-password">New Password (leave blank to keep current)</Label>
              <Input
                id="edit-password"
                type="password"
                value={editForm.password}
                onChange={(e) => setEditForm({ ...editForm, password: e.target.value })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button onClick={handleUpdateUser} disabled={submitting}>
              {submitting ? 'Saving...' : 'Save'}
            </Button>
            <Button variant="outline" onClick={() => setShowEditModal(false)}>
              Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      {dialogElement}
    </div>
  )
}
