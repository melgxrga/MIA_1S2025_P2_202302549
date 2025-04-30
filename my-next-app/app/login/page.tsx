'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function Login() {
  const [partition, setPartition] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ partition, username, password }),
      });
      const data = await res.json();
      if (res.ok && data.success) {
        localStorage.setItem('mia_session', JSON.stringify({ partition, username }));
        router.push('/disks');
      } else {
        setError(data.message || 'Usuario, contraseña o partición inválidos');
      }
    } catch (err) {
      setError('Error de conexión con el backend');
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen py-2">
      <h1 className="text-3xl font-bold mb-4">Login</h1>
      <form onSubmit={handleSubmit} className="bg-white p-8 rounded shadow-md w-80">
        <label className="block mb-2">ID Partición:</label>
        <input type="text" value={partition} onChange={e => setPartition(e.target.value)} className="mb-4 w-full border p-2 rounded" />
        <label className="block mb-2">Usuario:</label>
        <input type="text" value={username} onChange={e => setUsername(e.target.value)} className="mb-4 w-full border p-2 rounded" />
        <label className="block mb-2">Contraseña:</label>
        <input type="password" value={password} onChange={e => setPassword(e.target.value)} className="mb-4 w-full border p-2 rounded" />
        <button type="submit" className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded w-full">Submit</button>
        {error && <div className="text-red-500 mt-2">{error}</div>}
      </form>
    </div>
  );
}
