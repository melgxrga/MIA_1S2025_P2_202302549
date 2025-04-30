"use client";
import React from "react";

const API_URL = "http://localhost:8080";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

function Disks() {
  const router = useRouter();
  const [hasSession, setHasSession] = useState(true);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const session = localStorage.getItem("mia_session");
      if (!session) {
        setHasSession(false);
      }
    }
  }, []);

  if (!hasSession) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen py-2 relative">
        <button
          className="absolute top-6 right-8 px-6 py-2 bg-red-600 hover:bg-red-800 text-white font-bold rounded shadow-lg z-50"
          onClick={() => {
            localStorage.removeItem('mia_session');
            router.push('/login');
          }}
        >
          Cerrar Sesión
        </button>
        <h2 className="text-2xl font-bold mb-4 text-red-600">Acceso denegado</h2>
        <p>Debes iniciar sesión para ver el visualizador de archivos.</p>
        <button className="mt-4 px-4 py-2 bg-blue-500 hover:bg-blue-700 text-white rounded" onClick={() => router.push("/login")}>Ir a Login</button>
      </div>
    );
  }
  const [folder, setFolder] = useState("");
  const [disks, setDisks] = useState<any[]>([]);
  const [error, setError] = useState("");
  const [expandedDisk, setExpandedDisk] = useState<number | null>(null);

  const fetchDisks = async () => {
  setError("");
  setDisks([]);
  if (!folder) {
    setError("Debes ingresar la ruta de la carpeta");
    return;
  }
  try {
    const res = await fetch(`${API_URL}/disks?folder=${encodeURIComponent(folder)}`);
    if (!res.ok) {
      const err = await res.json();
      setError(err.error || "Error al consultar el backend");
      return;
    }
    const data = await res.json();
    setDisks(Array.isArray(data) ? data : []);
  } catch (e) {
    setError("Error al conectar con el backend");
  }
};


  console.log('DISKS:', disks);
  return (
    <div className="flex flex-col items-center justify-center min-h-screen py-2 relative">
      <button
        className="absolute top-6 right-8 px-6 py-2 bg-red-600 hover:bg-red-800 text-white font-bold rounded shadow-lg z-50"
        onClick={() => {
          localStorage.removeItem('mia_session');
          router.push('/login');
        }}
      >
        Cerrar Sesión
      </button>

      <h1 className="text-2xl font-bold mb-4">Visualizador de Discos</h1>
      <div className="flex gap-2 mb-4">
        <input
          type="text"
          placeholder="Ruta de la carpeta"
          value={folder}
          onChange={e => setFolder(e.target.value)}
          className="text-black border px-2 py-1 rounded"
          style={{ minWidth: 300 }}
        />
        <button
          onClick={fetchDisks}
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
        >
          Buscar discos
        </button>
      </div>
      {error && <div className="text-red-500 mb-2">{error}</div>}
      <div className="w-full flex flex-wrap gap-6 justify-center">
        {disks.map((disk, idx) => (
          <div
            key={idx}
            className={`border rounded p-4 bg-gray-50 shadow-md text-black cursor-pointer transition duration-150 hover:bg-gray-100 ${expandedDisk === idx ? 'ring-2 ring-blue-400' : ''}`}
            onClick={() => setExpandedDisk(expandedDisk === idx ? null : idx)}
          >
            <h2 className="font-bold text-lg mb-2">{disk.path.split("/").pop()}</h2>
            <div className="mb-1">Tamaño: {disk.size} bytes</div>
            <div className="mb-1">Fecha creación: {disk.creation_date}</div>
            <div className="mb-1">Signature: {disk.signature}</div>
            <div className="mb-1">Fit: {disk.fit}</div>
            <button
              className="mt-2 mb-1 px-3 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded font-semibold"
              onClick={e => {
                e.stopPropagation();
                window.location.href = `/partitions?disk=${encodeURIComponent(disk.path)}`;
              }}
            >
              Ver particiones
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}

export default Disks;
