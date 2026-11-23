import type { ReactNode } from "react";
import { CabinetShell } from "../../widgets/cabinet-shell";

export default function CabinetLayout({ children }: { children: ReactNode }) {
  return <CabinetShell>{children}</CabinetShell>;
}
