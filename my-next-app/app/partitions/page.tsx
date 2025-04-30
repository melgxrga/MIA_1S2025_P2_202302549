"use client";
import React, { useState } from "react";

const API_URL = "http://localhost:8080";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";

function Partitions() {
  const [diskPath, setDiskPath] = useState("");
  const [partitions, setPartitions] = useState<any[]>([]);
  const [error, setError] = useState("");
  const searchParams = typeof window !== "undefined" ? new URLSearchParams(window.location.search) : null;

  useEffect(() => {
    if (searchParams) {
      const disk = searchParams.get("disk");
      if (disk) {
        setDiskPath(disk);
        fetchPartitions(disk);
      }
    }
    // eslint-disable-next-line
  }, []);

  const fetchPartitions = async (overridePath?: string) => {
    setError("");
    setPartitions([]);
    const pathToUse = overridePath || diskPath;
    if (!pathToUse) {
      setError("Debes ingresar la ruta del disco");
      return;
    }
    try {
      const res = await fetch(`${API_URL}/partitions?disk=${encodeURIComponent(pathToUse)}`);
      if (!res.ok) {
        const err = await res.json();
        setError(err.error || "Error al consultar el backend");
        return;
      }
      const data = await res.json();
      setPartitions(Array.isArray(data) ? data : []);
    } catch (e) {
      setError("Error al conectar con el backend");
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen py-2">
      <h1 className="text-2xl font-bold mb-4">Explorador de Particiones</h1>
      {diskPath ? (
        <>
          <div className="flex flex-col items-center mb-6">
            <span className="text-lg font-semibold text-white mb-2">Disco: {diskPath.split("/").pop()}</span>
            <button
              onClick={() => window.location.href = "/disks"}
              className="bg-gray-600 hover:bg-gray-800 text-white px-4 py-2 rounded"
            >
              Volver a discos
            </button>
          </div>
          {error && (
            <div className="text-red-500 mb-2">
              {error}
              <br />
              <span className="text-xs text-gray-400">Debug: diskPath = {diskPath}</span>
              <br />
              <span className="text-xs text-gray-400">URL: {`http://localhost:8080/partitions?disk=${encodeURIComponent(diskPath)}`}</span>
            </div>
          )}
          <div className="w-full flex flex-wrap gap-6 justify-center">
            {partitions.length === 0 && <span className="text-gray-400">(Sin particiones)</span>}
            {partitions.map((part: any, idx: number) => (
              <div key={idx} className="border rounded px-4 py-2 bg-white text-black shadow-md min-w-[250px]">
                <div className="font-bold mb-1">Nombre: {part.name}</div>
                <div>Tipo: {String.fromCharCode(part.type)}</div>
                <div>Tamaño: {part.size} bytes</div>
                <div>Inicio: {part.start}</div>
                <div>Fit: {String.fromCharCode(part.fit)}</div>
                <div>Status: {String.fromCharCode(part.status)}</div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <div className="text-gray-400 mt-10">Selecciona un disco para ver sus particiones.</div>
      )}
    </div>
  );
}

export default Partitions;
