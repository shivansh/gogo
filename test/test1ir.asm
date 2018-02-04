	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0

	.text
main:
	la $t2, v1
	lw $t2, ($t2)
	li, $t2, -1
	la $t0, v2
	lw $t0, ($t0)
	li, $t0, 2
	la $t0, v3
	lw $t0, ($t0)
	li, $t0, 3
	la $t0, v4
	lw $t0, ($t0)
	li, $t0, 4
	la $t0, v5
	lw $t0, ($t0)
	li, $t0, 5
	jlt, $t0, 5, unsat
	la $t0, v2
	pingilw $t0, ($t0)
	move, $t0, $t2
	j temp

	; Store variables back into memory
	sw $t2, ($t2)
temp:
	jlt, $t0, $t0, unsat

	; Store variables back into memory
unsat:
	la $t2, v4
	lw $t2, ($t2)
	li, $t2, 4

	; Store variables back into memory
	sw $t2, ($t2)
