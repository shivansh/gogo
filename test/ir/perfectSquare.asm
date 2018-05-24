# Prints 1 if n is a perfect square

# Test to check whether a given number is a perfect square.

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
isStr:		.asciiz "n is not a perfect square."
notStr:		.asciiz "n is a perfect square."

	.text

	.globl main
	.ent main
main:
	li	$v0, 4
	la	$a0, nStr
	syscall
	li	$v0, 5
	syscall
	move	$3, $v0
	li	$7, 1		# i -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, i

loop:
	lw	$3, n
	lw	$7, i
	sub	$3, $3, $7	# n -> $3
	addi	$7, $7, 2	# i -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, i
	bgt	$3, 0, loop	# loop -> $0

	li	$v0, 4
	la	$a0, isStr
	syscall
	j	exit

	lw	$3, n
	# Store variables back into memory
	sw	$3, n
	bne	$3, 0, exit	# exit -> $0

	li	$v0, 4
	la	$a0, notStr
	syscall

exit:
	li	$v0, 10
	syscall
	.end main
