	.data
v1:	.word	0
v2:	.word	0
v4:	.word	0

	.text
main:
	la $t1, v1
	lw $t2, ($t1)
	li, $t2, -1
	li, $t2, 2
	jlt, $t2, 5, unsat
	la $t3, v2
	lw $t4, ($t3)
	move, $t4, $t2
	j temp

	; Store variables back into memory
	sw $t2, ($t1)
	sw $t4, ($t3)
temp:
	jlt, $t0, $t0, unsat

	; Store variables back into memory
unsat:
	la $t1, v4
	lw $t2, ($t1)
	li, $t2, 4

	; Store variables back into memory
	sw $t2, ($t1)
