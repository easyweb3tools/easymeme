import { getDefaultConfig } from '@rainbow-me/rainbowkit';
import { bsc } from 'wagmi/chains';

export const config = getDefaultConfig({
  appName: 'EasyMeme',
  projectId: process.env.NEXT_PUBLIC_WALLET_CONNECT_ID || 'demo',
  chains: [bsc],
  ssr: true,
});
