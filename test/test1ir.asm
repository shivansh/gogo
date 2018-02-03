	.data
v1:	.word	0
v2:	.word	0
v4:	.word	0

	.text
main:
	; Load variables from memory into registers
	la $t1, v1
	lw $t2, ($t1)
	la $t3, v2
	lw $t4, ($t3)

	li, $t5, -1
	jlt, $t5, 5, unsat
	move, $t6, $t5
	j temp

	; Store variables back into memory
temp:
	; Load variables from memory into registers
	la $t1, v1
	lw $t2, ($t1)
	la $t3, v2
	lw $t4, ($t3)

	jlt, $t2, $t4, unsat

	; Store variables back into memory
	sw $t2, ($t1)
	sw $t4, ($t3)
unsat:
	; Load variables from memory into registers
	la $t1, v1
	lw $t2, ($t1)
	la $t3, v2
	lw $t4, ($t3)
	la $t5, v4
	lw $t6, ($t5)

	li, $t7, 4

	; Store variables back into memory
	sw $t2, ($t1)
	sw $t4, ($t3)
