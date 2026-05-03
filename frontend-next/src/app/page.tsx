export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="z-10 max-w-5xl w-full items-center justify-center font-mono text-sm">
        <h1 className="text-4xl font-bold mb-4">GODNSLOG 2.0</h1>
        <p className="text-xl mb-8">OAST交互验证与证据平台</p>
        <div className="grid gap-4 text-center">
          <a href="/login" className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90">
            登录
          </a>
        </div>
      </div>
    </main>
  )
}
