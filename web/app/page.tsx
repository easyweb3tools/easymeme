import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-b from-background to-muted">
      <h1 className="text-5xl font-bold mb-4">EasyMeme</h1>
      <p className="text-xl text-muted-foreground mb-8">
        AI-powered BNB Chain Meme Token Scanner
      </p>
      <Link
        href="/dashboard"
        className="px-8 py-4 bg-primary text-primary-foreground rounded-lg text-lg font-bold"
      >
        Launch App
      </Link>
    </div>
  );
}
