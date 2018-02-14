# Test to check whether a given number is a perfect square.

	.data
n:	.word	0
i:	.word	0
r:	.word	0

	.text


	.globl main
	.ent main
main:
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 1		# i -> $t4

	# Store variables back into memory
	sw $t4, i
	sw $t1, n

loop:
	lw $t1, n
	lw $t4, i
	sub $t1, $t1, $t4	# n -> $t1
	addi $t4, $t4, 2	# i -> $t4

	# Store variables back into memory
	sw $t1, n
	sw $t4, i

	lw $t1, n
	bgt $t1, 0, loop		# loop -> $t0
	li $t4, 0		# r -> $t4
	# Prints 1 if n is a perfect square

	# Store variables back into memory
	sw $t1, n
	sw $t4, r

	lw $t1, n
	bne $t1, 0, exit		# exit -> $t0
	li $t4, 1		# r -> $t4

	# Store variables back into memory
	sw $t1, n
	sw $t4, r

exit:
	li $v0, 1
	lw $t1, r
	move $a0, $t1
	syscall

	# Store variables back into memory
	sw $t1, r
	li $v0, 10
	syscall
	.end main
