import { ConnectButton } from '@rainbow-me/rainbowkit';
import { TokenList } from '@/components/token-list';

export default function DashboardPage() {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold">EasyMeme</h1>
          <ConnectButton />
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <TokenList />
      </main>
    </div>
  );
}
