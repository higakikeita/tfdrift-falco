import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { GoogleAnalytics } from '@next/third-parties/google'
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "TFDrift-Falco - Real-time Terraform Drift Detection",
  description: "Detect manual infrastructure changes instantly using Falco's CloudTrail plugin. Monitor 120+ AWS CloudTrail events across 12 services with real-time alerts.",
  keywords: ["terraform", "drift detection", "falco", "cloudtrail", "aws", "infrastructure", "devops", "security"],
  authors: [{ name: "Keita Higaki", url: "https://github.com/higakikeita" }],
  openGraph: {
    title: "TFDrift-Falco - Real-time Terraform Drift Detection",
    description: "Detect manual infrastructure changes instantly using Falco's CloudTrail plugin",
    url: "https://tfdrift-falco.vercel.app",
    siteName: "TFDrift-Falco",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "TFDrift-Falco - Real-time Terraform Drift Detection",
    description: "Detect manual infrastructure changes instantly using Falco's CloudTrail plugin",
    creator: "@keitah0322",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
        {process.env.NEXT_PUBLIC_GA_ID && (
          <GoogleAnalytics gaId={process.env.NEXT_PUBLIC_GA_ID} />
        )}
      </body>
    </html>
  );
}
