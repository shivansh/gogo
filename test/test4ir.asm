	.data
v1:	.word	0
v2:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0
v6:	.word	0

	.text


	.globl main
	.ent main
main:
	li $t1, 1		# v1 -> $t1
	li $t4, 2		# v2 -> $t4
	li $t3, 5		# v3 -> $t3
	li $t2, 4		# v4 -> $t2
	sw $t3, v3		# spilled v3, freed $t3
	add $t3, $t1, $t4	# v5 -> $t3
	sw $t3, v5		# spilled v5, freed $t3
	lw $t3, v3
	sw $t2, v4		# spilled v4, freed $t2
	add $t2, $t3, $t2	# v6 -> $t2
	li $v0, 1
	sw $t3, v3		# spilled v3, freed $t3
	lw $t3, v5
	move $a0, $t3
	syscall

	# Store variables back into memory
	sw $t1, v1
	sw $t4, v2
	sw $t3, v5
	sw $t2, v6
	li $v0, 10
	syscall
	.end main
