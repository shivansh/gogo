	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0

	.text
main:
	la $t1, v1
	li, $t2, -1		; v1 -> {reg: $t2, mem: $t1}
	la $t3, v2
	li, $t4, 2		; v2 -> {reg: $t4, mem: $t3}
	li, $t2, -12
	sw $t2, ($t1)		; spilled v1 and freed {$t2, $t1}
	la $t2, v3
	li, $t1, 3		; v3 -> {reg: $t1, mem: $t2}
	la $t2, v4
	sw $t1, ($t2)		; spilled v3 and freed {$t1, $t2}
	li, $t1, 4		; v4 -> {reg: $t1, mem: $t2}
	la $t1, v5
	li, $t2, 5		; v5 -> {reg: $t2, mem: $t1}
	jlt, $t0, 5, unsat
	j temp

	; Store variables back into memory
	sw $t2, ($t1)
	sw $t4, ($t3)
	sw $t1, ($t2)
temp:
	jlt, $t0, $t0, unsat

	; Store variables back into memory
unsat:
	la $t1, v4
	li, $t2, 4		; v4 -> {reg: $t2, mem: $t1}

	; Store variables back into memory
	sw $t2, ($t1)
