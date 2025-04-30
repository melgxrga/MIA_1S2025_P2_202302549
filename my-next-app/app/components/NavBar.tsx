"use client";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";

export default function NavBar() {
  const [hasSession, setHasSession] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (typeof window !== "undefined") {
      setHasSession(!!localStorage.getItem("mia_session"));
    }
  }, [pathname]);

  const handleDisksClick = (e: React.MouseEvent) => {
    if (!localStorage.getItem("mia_session")) {
      e.preventDefault();
      router.push("/login?redirect=/disks");
    }
  };

  return (
    <nav className="w-full flex gap-8 px-8 py-4 bg-gray-200 mb-4">
      <Link href="/" className="font-bold text-blue-700 hover:underline">Terminal</Link>
      <Link href="/login" className="font-bold text-blue-700 hover:underline">
        Visualizador de Archivos
      </Link>
    </nav>
  );
}
