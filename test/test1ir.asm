	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0

	.text
main:
	la $t4, v1
	li, $t1, -1		; v1 -> {reg: $t1, mem: $t4}
	la $t2, v2
	li, $t3, 2		; v2 -> {reg: $t3, mem: $t2}
	li, $t1, -12
	sw $t1, ($t4)		; spilled v1 and freed {$t1, $t4}
	la $t1, v3
	li, $t4, 3		; v3 -> {reg: $t4, mem: $t1}
	sw $t4, ($t1)		; spilled v3 and freed {$t4, $t1}
	la $t4, v4
	li, $t1, 4		; v4 -> {reg: $t1, mem: $t4}
	sw $t1, ($t4)		; spilled v4 and freed {$t1, $t4}
	la $t1, v5
	li, $t4, 5		; v5 -> {reg: $t4, mem: $t1}
	jlt, $t0, 5, unsat
	move, $t3, $t0
	j temp

	; Store variables back into memory
	sw $t4, ($t1)
	sw $t3, ($t2)
temp:
	jlt, $t0, $t0, unsat

	; Store variables back into memory
unsat:
	la $t1, v4
	li, $t2, 4		; v4 -> {reg: $t2, mem: $t1}

	; Store variables back into memory
	sw $t2, ($t1)
