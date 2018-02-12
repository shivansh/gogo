	.data
n:	.word	0
i:	.word	0
k:	.word	0

	.text


	.globl main
	.ent main
main:
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 0		# i -> $t4
	sw $t1, n		# spilled n, freed $t1
	li $t1, 0		# k -> $t1

	# Store variables back into memory
	sw $t4, i
	sw $t1, k

loop:

	lw $t1, i
	lw $t4, n
	bge $t1, $t4, exit		# exit -> $t0
	lw $t3, k
	add $t3, $t3, $t1	# k -> $t3
	addi $t1, $t1, 1		# i -> $t1

	# Store variables back into memory
	sw $t1, i
	sw $t4, n
	sw $t3, k

	j loop

exit:
	li $v0, 1
	lw $t1, k
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, k
	li $v0, 10
	syscall
	.end main
