import { Check, UserPlus } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ownerChoices, toggleOwner } from '@/lib/assigns'

export function OwnerPicker({
  people,
  owners,
  saving,
  readOnly,
  onChange,
}: {
  people: string[]
  owners: string[]
  saving: boolean
  readOnly: boolean
  onChange: (owners: string[]) => void
}) {
  const [open, setOpen] = useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="xs"
          disabled={saving || readOnly}
          aria-label="Pick owners"
          className="text-muted-foreground"
        >
          <UserPlus className="size-3.5" />
          {saving && 'Saving…'}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-56 p-0">
        <Command>
          <CommandInput placeholder="Find a person…" />
          <CommandList>
            <CommandEmpty>No one found.</CommandEmpty>
            {ownerChoices(people, owners).map((name) => (
              <CommandItem
                key={name}
                value={name}
                onSelect={() => {
                  setOpen(false)
                  onChange(toggleOwner(owners, name))
                }}
              >
                <span className="min-w-0 flex-1 truncate">{name}</span>
                {owners.includes(name) && <Check className="size-3.5" />}
              </CommandItem>
            ))}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
