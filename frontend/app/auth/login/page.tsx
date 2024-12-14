import { LoginForm } from '@/components/auth/login-form'
import { ModeToggle } from '@/components/mode-toggle'

export default function LoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <ModeToggle />
      <LoginForm />
    </div>
  )
}
