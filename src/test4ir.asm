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
	li $t2, 5		# v3 -> $t2
	sw $t4, v2		# spilled v2, freed $t4
	li $t4, 4		# v4 -> $t4
	addi $t3, $t1, 3		# v5 -> $t3
	sw $t4, v4		# spilled v4, freed $t4
	add $t4, $t2, $t3	# v6 -> $t4
	mul $t4, $t4, 5		# v6 -> $t4

	# Store variables back into memory
	sw $t3, v5
	sw $t1, v1
	sw $t4, v6
	sw $t2, v3

	lw $t1, v5
	lw $t4, v6
	bgt $t1, $t4, temp		# temp -> $t0
	sub $t4, $t4, $t1	# v6 -> $t4
	li $v0, 1
	move $a0, $t4
	syscall

	# Store variables back into memory
	sw $t1, v5
	sw $t4, v6

temp:
	li $v0, 1
	lw $t1, v6
	move $a0, $t1
	syscall
	sll $t1, $t1, 2		# v6 -> $t1
	li $v0, 1
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, v6
	li $v0, 10
	syscall
	.end main
