	.data
v7:	.word	0
v1:	.word	0
v3:	.word	0
v4:	.word	0
v5:	.word	0

	.text


	.globl main
	.ent main
main:
	li $t1, 7		# v7 -> $t1
	li $t4, 1		# v1 -> $t4

	# Store variables back into memory
	sw $t4, v1
	sw $t1, v7

	lw $t1, v7
	bgt $t1, 1, temp		# temp -> $t0

	# Store variables back into memory
	sw $t1, v7

temp1:
	li $t1, 3		# v3 -> $t1
	li $t4, 4		# v4 -> $t4
	lw $t2, v1
	sw $t4, v4		# spilled v4, freed $t4
	add $t4, $t2, $t1	# v5 -> $t4
	li $v0, 1
	move $a0, $t4
	syscall

	# Store variables back into memory
	sw $t2, v1
	sw $t1, v3
	sw $t4, v5

temp:
	lw $t1, v1
	addi $t1, $t1, 4		# v1 -> $t1

	# Store variables back into memory
	sw $t1, v1

	lw $t1, v1
	ble $t1, 13, temp1		# temp1 -> $t0

	# Store variables back into memory
	sw $t1, v1

	j exit
	li $t1, 100		# v1 -> $t1

	# Store variables back into memory
	sw $t1, v1

exit:
	li $t1, 105		# v1 -> $t1
	li $v0, 1
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, v1
	li $v0, 10
	syscall
	.end main
