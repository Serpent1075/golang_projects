import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"


export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function apiRequest(path: string) {

  fetch(`http://localhost:3010/v1/${path}`,{
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'

    },
    body: JSON.stringify({
      "name": "John Doe",
    }),
  },
  )
    .then(response => response.json())
    .then(data => console.log(data))
    .catch(error => console.log(error));
}