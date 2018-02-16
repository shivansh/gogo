# Test to check whether a given number is a perfect square.

	.data
nStr:	.asciiz "Enter n: "
n:	.word	0
i:	.word	0
isStr:	.asciiz "n is not a perfect square."
notStr:	.asciiz "n is a perfect square."

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, nStr
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 1		# i -> $t4
	# Store variables back into memory
	sw $t1, n
	sw $t4, i

loop:
	lw $t1, n
	lw $t4, i
	sub $t1, $t1, $t4	# n -> $t1
	addi $t4, $t4, 2		# i -> $t4
	# Store variables back into memory
	sw $t1, n
	sw $t4, i

	lw $t1, n
	bgt $t1, 0, loop		# loop -> $t0
	li $v0, 4
	la $a0, isStr
	syscall
	# Store variables back into memory
	sw $t1, n

	j exit
	# Prints 1 if n is a perfect square

	lw $t1, n
	bne $t1, 0, exit		# exit -> $t0
	li $v0, 4
	la $a0, notStr
	syscall
	# Store variables back into memory
	sw $t1, n

exit:
	li $v0, 10
	syscall
	.end main
