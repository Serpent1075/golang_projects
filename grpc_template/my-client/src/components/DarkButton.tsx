import { cn } from '@/lib/utils'
import React, { MouseEventHandler } from 'react'

const DarkButton = ({icon = null, label, className, onClick, ...props}:{icon:any, label:string, className:string, onClick:MouseEventHandler<HTMLDivElement>}) => {
  return (
    <div 
    onClick={onClick}
    className={cn(
        "cursor-pointer border border-neutral-700 bg-black hover:bg-neutral-700 text-white rounded-2xl flex flex-row items-center min-w-[80px] h-[36px] p-4 gap-2",
        className
    )}
    {...props}>
        <span>{icon}</span>
        <span>{label}</span>
    </div>
  )
}


export default DarkButton