# Test to find n'th fibonacci number

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
retVal:		.word	0
str:		.asciiz "n'th fibonacci number: "
first:		.word	0
second:		.word	0
temp:		.word	0

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
	li	$7, 0		# i -> $7
	sw	$3, n
	sw	$7, i
	jal	fib
	lw	$3, n
	lw	$7, i
	move	$15, $v0
	li	$v0, 4
	la	$a0, str
	syscall
	li	$v0, 1
	move	$a0, $15
	syscall
	# Store variables back into memory
	sw	$3, n
	sw	$7, i
	sw	$15, retVal
	li	$v0, 10
	syscall
	.end main
	.globl fib
	.ent fib
fib:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$3, 1		# first -> $3
	li	$7, 1		# second -> $7
	lw	$15, n
	sub	$15, $15, 2	# n -> $15
	# Store variables back into memory
	sw	$3, first
	sw	$7, second
	sw	$15, n

loop:
	lw	$3, n
	# Store variables back into memory
	sw	$3, n
	ble	$3, 0, exit	# exit -> $0

	lw	$3, second
	move	$7, $3		# temp -> $7
	lw	$30, first
	add	$3, $3, $30	# second -> $3
	move	$30, $7		# first -> $30
	lw	$15, n
	sub	$15, $15, 1	# n -> $15
	# Store variables back into memory
	sw	$3, second
	sw	$7, temp
	sw	$15, n
	sw	$30, first
	j	loop

exit:
	lw	$v0, second

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end fib
