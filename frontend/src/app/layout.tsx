import type { Metadata } from "next";
import type { ReactNode } from "react";

import { Providers } from "./providers";
import "./styles/tokens.scss";
import "./styles/globals.scss";

export const metadata: Metadata = {
  title: "Transfer Service",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
